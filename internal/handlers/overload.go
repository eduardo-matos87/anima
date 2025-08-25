package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
)

type overloadReq struct {
	ExercicioID int64 `json:"exercicio_id"`
	Window      int   `json:"window,omitempty"` // janela de sessões recentes
}
type overloadResp struct {
	SuggestedCargaKg float64 `json:"suggested_carga_kg"`
	SuggestedReps    int     `json:"suggested_repeticoes"`
	Rationale        string  `json:"rationale"`
}

// POST /api/overload/suggest
// Regra simples:
// - média das últimas N cargas do exercício
// - se média RIR >= 1.8 => +2.5kg; se <= 0.5 => reduzir reps
func OverloadSuggest(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in overloadReq
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ExercicioID <= 0 {
			badRequest(w, "invalid exercicio_id")
			return
		}
		if in.Window <= 0 || in.Window > 12 {
			in.Window = 5
		}
		row := db.QueryRow(`
		   SELECT COALESCE(AVG(carga_kg),0), COALESCE(AVG(rir),1.5)
		   FROM workout_sets
		   WHERE exercicio_id = $1
		   ORDER BY id DESC
		   LIMIT $2
		`, in.ExercicioID, in.Window)
		var avgCarga, avgRIR float64
		if err := row.Scan(&avgCarga, &avgRIR); err != nil {
			internalErr(w, err)
			return
		}
		resp := overloadResp{
			SuggestedCargaKg: roundTo(avgCarga, 0.5),
			SuggestedReps:    10,
			Rationale:        "baseado na média recente",
		}
		if avgRIR >= 1.8 {
			resp.SuggestedCargaKg = roundTo(avgCarga+2.5, 0.5)
			resp.Rationale = "RIR alto, sugere aumentar 2.5kg"
		} else if avgRIR <= 0.5 {
			resp.SuggestedReps = 8
			resp.Rationale = "RIR baixo, manter carga e reduzir reps"
		}
		jsonWrite(w, http.StatusOK, resp)
	})
}

func roundTo(v, step float64) float64 {
	return math.Round(v/step) * step
}

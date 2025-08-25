package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

type overloadReq struct {
	ExercicioID int64 `json:"exercicio_id"`
	Window      int   `json:"window,omitempty"` // 3..12 (default 5)
}
type overloadResp struct {
	SuggestedCargaKg float64 `json:"suggested_carga_kg"`
	SuggestedReps    int     `json:"suggested_repeticoes"`
	Rationale        string  `json:"rationale"`
	AvgCargaKg       float64 `json:"avg_carga_kg"`
	AvgRIR           float64 `json:"avg_rir"`
	SampleCount      int     `json:"sample_count"`
}

// POST /api/overload/suggest  (body JSON)
// GET  /api/overload/suggest?exercicio_id=10&window=5
// GET  /api/suggestions/next-load?exercicio_id=10&window=5  (legacy via compat)
func OverloadSuggest(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in overloadReq

		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			idStr := q.Get("exercicio_id")
			if idStr == "" {
				badRequest(w, "missing exercicio_id")
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil || id <= 0 {
				badRequest(w, "invalid exercicio_id")
				return
			}
			in.ExercicioID = id
			if win := q.Get("window"); win != "" {
				if wv, err := strconv.Atoi(win); err == nil {
					in.Window = wv
				}
			}
		default: // POST
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ExercicioID <= 0 {
				badRequest(w, "invalid json or exercicio_id")
				return
			}
		}

		// saneamento da janela
		if in.Window < 3 || in.Window > 12 {
			in.Window = 5
		}

		var (
			avgCarga float64
			avgRIR   float64
			n        int
		)

		// Se window == 12, usamos a view otimizada (migration 023)
		if in.Window == 12 {
			// As views usam NUMERIC; convertemos para float8 no SELECT
			row := db.QueryRow(`
				SELECT
					COALESCE(avg_carga_kg::float8, 0),
					COALESCE(avg_rir::float8, 1.5),
					COALESCE(sample_count, 0)
				FROM workout_overload_stats12
				WHERE exercicio_id = $1
			`, in.ExercicioID)
			if err := row.Scan(&avgCarga, &avgRIR, &n); err != nil {
				internalErr(w, err)
				return
			}
		} else {
			// Subquery: últimos N concluídos para o exercício
			row := db.QueryRow(`
			   SELECT
			     COALESCE(AVG(carga_kg), 0) AS avg_carga,
			     COALESCE(AVG(rir), 1.5)    AS avg_rir,
			     COUNT(*)                    AS n
			   FROM (
			     SELECT carga_kg, rir
			     FROM workout_sets
			     WHERE exercicio_id = $1
			       AND completed = TRUE
			     ORDER BY id DESC
			     LIMIT $2
			   ) s
			`, in.ExercicioID, in.Window)
			if err := row.Scan(&avgCarga, &avgRIR, &n); err != nil {
				internalErr(w, err)
				return
			}
		}

		// sem amostras: resposta neutra
		if n == 0 {
			jsonWrite(w, http.StatusOK, overloadResp{
				SuggestedCargaKg: 0,
				SuggestedReps:    10,
				Rationale:        "sem histórico concluído para este exercício",
				AvgCargaKg:       0,
				AvgRIR:           1.5,
				SampleCount:      0,
			})
			return
		}

		// regra adaptativa
		sugCarga := roundTo(avgCarga, 0.5)
		sugReps := 10
		rationale := "RIR moderado, manter carga"

		switch {
		// +5 kg quando a folga é grande
		case avgRIR >= 2.5:
			sugCarga = roundTo(avgCarga+5.0, 0.5)
			rationale = "RIR muito alto, sugere +5kg"

			// +2.5 kg quando ainda há margem
		case avgRIR >= 1.8:
			sugCarga = roundTo(avgCarga+2.5, 0.5)
			rationale = "RIR alto, sugere +2.5kg"

			// reduzir reps quando está no limite
		case avgRIR <= 0.5:
			sugReps = 8
			rationale = "RIR baixo, manter carga e reduzir reps"
		}

		jsonWrite(w, http.StatusOK, overloadResp{
			SuggestedCargaKg: sugCarga,
			SuggestedReps:    sugReps,
			Rationale:        rationale,
			AvgCargaKg:       roundTo(avgCarga, 0.5),
			AvgRIR:           avgRIR,
			SampleCount:      n,
		})
	})
}

func roundTo(v, step float64) float64 {
	return math.Round(v/step) * step
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type nextLoadResp struct {
	ExercicioID   int64    `json:"exercicio_id"`
	BasedOnSetID  *int64   `json:"based_on_set_id,omitempty"`
	CurrentWeight *float64 `json:"current_weight_kg,omitempty"`
	CurrentReps   *int     `json:"current_reps,omitempty"`
	CurrentRIR    *int     `json:"current_rir,omitempty"`
	SuggestWeight *float64 `json:"suggest_weight_kg,omitempty"`
	SuggestReps   *int     `json:"suggest_reps,omitempty"`
	Note          string   `json:"note"`
}

func roundTo(v float64, step float64) float64 {
	if step <= 0 {
		return v
	}
	return math.Round(v/step) * step
}

func NextLoad(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uid := getUserID(r)

		q := r.URL.Query()
		exIDStr := q.Get("exercicio_id")
		if strings.TrimSpace(exIDStr) == "" {
			http.Error(w, "exercicio_id required", http.StatusBadRequest)
			return
		}
		exID, err := strconv.ParseInt(exIDStr, 10, 64)
		if err != nil || exID <= 0 {
			http.Error(w, "invalid exercicio_id", http.StatusBadRequest)
			return
		}
		recent := 5
		if v := q.Get("recent_sessions"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 20 {
				recent = n
			}
		}

		// Busca últimos sets do usuário para o exercício, ordenados por data (mais recente primeiro)
		const sqlq = `
SELECT ws.id, ws.weight_kg, ws.reps, ws.rir
FROM workout_sets ws
JOIN workout_sessions s ON s.id = ws.session_id
WHERE s.user_id = $1 AND ws.exercicio_id = $2
ORDER BY ws.created_at DESC
LIMIT $3`
		rows, err := db.Query(sqlq, uid, exID, recent*6) // até ~6 sets por sessão
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type row struct {
			id    int64
			kg    *float64
			reps  *int
			rir   *int
			score float64
		}
		var history []row
		for rows.Next() {
			var (
				id   int64
				kg   sql.NullFloat64
				reps sql.NullInt64
				rir  sql.NullInt64
			)
			if err := rows.Scan(&id, &kg, &reps, &rir); err != nil {
				http.Error(w, "scan error", http.StatusInternalServerError)
				return
			}
			var h row
			h.id = id
			if kg.Valid {
				v := kg.Float64
				h.kg = &v
			}
			if reps.Valid {
				v := int(reps.Int64)
				h.reps = &v
			}
			if rir.Valid {
				v := int(rir.Int64)
				h.rir = &v
			}
			if h.kg != nil && h.reps != nil {
				h.score = (*h.kg) * float64(*h.reps)
			}
			history = append(history, h)
		}

		resp := nextLoadResp{ExercicioID: exID}
		if len(history) == 0 {
			resp.Note = "Sem histórico: realize um set de teste para calibrar (ex.: 8–12 reps com carga confortável)."
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Escolhe referência: maior “tonelagem”; se empatar, preferir maior RIR.
		bestIdx := -1
		bestScore := -1.0
		for i, h := range history {
			if h.kg == nil || h.reps == nil {
				continue
			}
			score := h.score
			// leve viés por RIR (cada RIR soma +1% na pontuação)
			if h.rir != nil {
				score = score * (1.0 + 0.01*float64(*h.rir))
			}
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}
		if bestIdx < 0 {
			resp.Note = "Histórico sem peso/reps válidos — faça um set válido para calibrar."
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(resp)
			return
		}

		ref := history[bestIdx]
		resp.BasedOnSetID = &ref.id
		resp.CurrentWeight = ref.kg
		resp.CurrentReps = ref.reps
		resp.CurrentRIR = ref.rir

		step := 0.5 // arredondamento de anilha (ajuste se usar 1.25 etc.)
		var suggestKg *float64
		var suggestReps *int
		note := "Baseado no melhor set recente."

		get := func() (float64, int, int) {
			kg := 0.0
			if ref.kg != nil {
				kg = *ref.kg
			}
			rp := 0
			if ref.reps != nil {
				rp = *ref.reps
			}
			rr := 0
			if ref.rir != nil {
				rr = *ref.rir
			}
			return kg, rp, rr
		}

		kg, rp, rr := get()
		switch {
		case rr >= 3:
			n := roundTo(kg*1.05, step)
			suggestKg = &n
			r := rp
			suggestReps = &r
			note = "RIR≥3: sugerido +5% de carga."
		case rr == 2:
			n := roundTo(kg*1.025, step)
			suggestKg = &n
			r := rp
			suggestReps = &r
			note = "RIR=2: sugerido +2.5% de carga."
		case rr == 1:
			n := roundTo(kg, step)
			suggestKg = &n
			r := rp + 1
			suggestReps = &r
			note = "RIR=1: manter carga e tentar +1 rep."
		default: // rr <= 0, ou sem RIR informado
			if rr <= 0 {
				n := roundTo(kg*0.975, step)
				suggestKg = &n
				r := rp
				if r > 1 {
					r = rp - 1
				}
				suggestReps = &r
				note = "RIR≤0: reduzir ~2.5% ou -1 rep para recuperar."
			} else {
				// rr nil — ser conservador
				n := roundTo(kg, step)
				suggestKg = &n
				r := rp
				suggestReps = &r
				note = "Sem RIR: manter carga; foque técnica e progressão gradual."
			}
		}

		resp.SuggestWeight = suggestKg
		resp.SuggestReps = suggestReps
		resp.Note = note

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(resp)
	})
}

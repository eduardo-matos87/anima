package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Body esperado:
//
//	{
//	  "items": [
//	    { "id": 4, "weight_kg": 45, "reps": 9, "rir": 1, "completed": true, "notes": "opcional" },
//	    { "id": 5, "carga_kg": 47.5, "repeticoes": 8 }  // suporte legado
//	  ]
//	}
//
// Resposta: { "updated": N, "failed": [ids...], "total": M }
func SetsBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	db := sessionsDB
	if db == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}

	var in struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || len(in.Items) == 0 {
		badRequest(w, "invalid json or empty items")
		return
	}

	userID := strings.TrimSpace(GetUserID(r))

	updated := 0
	failed := make([]int64, 0, len(in.Items))

	for _, it := range in.Items {
		// id obrigatório
		idAny, ok := it["id"]
		if !ok {
			failed = append(failed, 0)
			continue
		}
		id, ok := toInt64(idAny)
		if !ok || id <= 0 {
			failed = append(failed, 0)
			continue
		}

		// permitir nomes novos e legados
		// mapeia para colunas atuais: weight_kg, reps, rir, completed, notes, rest_sec
		setParts := []string{}
		args := []any{}
		argIdx := 1

		for k, v := range it {
			switch k {
			case "id":
				// ignora
			case "weight_kg", "carga_kg": // novo e legado → weight_kg
				fv, ok := toFloat64(v)
				if !ok {
					badRequest(w, "weight_kg/carga_kg must be number")
					return
				}
				setParts = append(setParts, "weight_kg = $"+itoa(argIdx))
				args = append(args, fv)
				argIdx++
			case "reps", "repeticoes": // novo e legado → reps
				iv, ok := toInt64(v)
				if !ok {
					badRequest(w, "reps/repeticoes must be int")
					return
				}
				setParts = append(setParts, "reps = $"+itoa(argIdx))
				args = append(args, iv)
				argIdx++
			case "rir":
				iv, ok := toInt64(v)
				if !ok {
					badRequest(w, "rir must be int")
					return
				}
				setParts = append(setParts, "rir = $"+itoa(argIdx))
				args = append(args, iv)
				argIdx++
			case "completed":
				b, ok := v.(bool)
				if !ok {
					badRequest(w, "completed must be bool")
					return
				}
				setParts = append(setParts, "completed = $"+itoa(argIdx))
				args = append(args, b)
				argIdx++
			case "notes":
				s, ok := v.(string)
				if !ok {
					badRequest(w, "notes must be string")
					return
				}
				setParts = append(setParts, "notes = $"+itoa(argIdx))
				args = append(args, s)
				argIdx++
			case "rest_sec":
				iv, ok := toInt64(v)
				if !ok {
					badRequest(w, "rest_sec must be int")
					return
				}
				setParts = append(setParts, "rest_sec = $"+itoa(argIdx))
				args = append(args, iv)
				argIdx++
			default:
				// ignora chaves desconhecidas
			}
		}

		if len(setParts) == 0 {
			// nada pra atualizar nesse item
			failed = append(failed, id)
			continue
		}

		q := `
UPDATE workout_sets AS s
SET ` + strings.Join(setParts, ", ") + `
FROM workout_sessions AS ws
WHERE s.id = $` + itoa(argIdx) + `
  AND ws.id = s.session_id
  AND ($` + itoa(argIdx+1) + ` = '' OR ws.user_id IS NULL OR ws.user_id = $` + itoa(argIdx+1) + `)
`
		args = append(args, id, userID)

		res, err := db.Exec(q, args...)
		if err != nil {
			internalErr(w, err)
			return
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			failed = append(failed, id)
		} else {
			updated++
		}
	}

	jsonWrite(w, http.StatusOK, map[string]any{
		"updated": updated,
		"failed":  failed,
		"total":   len(in.Items),
	})
}

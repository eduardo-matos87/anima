package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// PATCH /api/sets/batch
// Body: { "items": [ { "id": 1, "completed": true, "rir": 2, "carga_kg": 55, "repeticoes": 10, "notes": "ok" }, ... ] }
func SetsBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(GetUserID(r))

	var in struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if len(in.Items) == 0 {
		badRequest(w, "items required")
		return
	}

	tx, err := sessionsDB.Begin()
	if err != nil {
		internalErr(w, err)
		return
	}
	defer tx.Rollback()

	updated := 0
	for _, it := range in.Items {
		idAny, ok := it["id"]
		if !ok {
			badRequest(w, "each item must have id")
			return
		}
		id, ok := toInt64(idAny)
		if !ok || id <= 0 {
			badRequest(w, "invalid set id")
			return
		}

		// Campos permitidos
		allowed := map[string]bool{
			"completed":  true,
			"rir":        true,
			"carga_kg":   true,
			"repeticoes": true,
			"notes":      true,
		}
		setParts := []string{}
		args := []any{}
		argIdx := 1

		for k, v := range it {
			if k == "id" {
				continue
			}
			if !allowed[k] {
				badRequest(w, "unsupported field: "+k)
				return
			}
			switch k {
			case "completed":
				b, ok := v.(bool)
				if !ok {
					badRequest(w, "completed must be bool")
					return
				}
				setParts = append(setParts, "completed = $"+itoa(argIdx))
				args = append(args, b)
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
			case "carga_kg":
				fv, ok := toFloat64(v)
				if !ok {
					badRequest(w, "carga_kg must be number")
					return
				}
				setParts = append(setParts, "carga_kg = $"+itoa(argIdx))
				args = append(args, fv)
				argIdx++
			case "repeticoes":
				iv, ok := toInt64(v)
				if !ok {
					badRequest(w, "repeticoes must be int")
					return
				}
				setParts = append(setParts, "repeticoes = $"+itoa(argIdx))
				args = append(args, iv)
				argIdx++
			case "notes":
				sv, ok := v.(string)
				if !ok {
					badRequest(w, "notes must be string")
					return
				}
				setParts = append(setParts, "notes = $"+itoa(argIdx))
				args = append(args, sv)
				argIdx++
			}
		}

		if len(setParts) == 0 {
			continue
		}

		// UPDATE com guarda de dono (user_id). Aceita sessões antigas com user_id NULL.
		q := `
UPDATE workout_sets AS s
SET ` + strings.Join(setParts, ", ") + `
FROM workout_sessions AS ws
WHERE s.id = $` + itoa(argIdx) + `
  AND ws.id = s.session_id
  AND ($` + itoa(argIdx+1) + ` = '' OR ws.user_id IS NULL OR ws.user_id = $` + itoa(argIdx+1) + `)
`
		args = append(args, id, userID)

		res, err := tx.Exec(q, args...)
		if err != nil {
			internalErr(w, err)
			return
		}
		if aff, _ := res.RowsAffected(); aff > 0 {
			updated++
		}
	}

	if err := tx.Commit(); err != nil {
		internalErr(w, err)
		return
	}
	jsonWrite(w, http.StatusOK, map[string]any{"updated": updated})
}

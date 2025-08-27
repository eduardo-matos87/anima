package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// GET /api/sessions/sets/{session_id}
// POST /api/sessions/sets/{session_id}
func SessionSetsCollection(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		// /api /sessions /sets /{session_id}
		if len(parts) < 4 {
			badRequest(w, "missing session id")
			return
		}
		sessionID, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil || sessionID <= 0 {
			badRequest(w, "invalid session id")
			return
		}
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query(`
			   SELECT id, session_id, exercicio_id, series, repeticoes, COALESCE(carga_kg,0), COALESCE(rir,0), completed, COALESCE(notes,'')
			   FROM workout_sets WHERE session_id = $1 ORDER BY id ASC`, sessionID)
			if err != nil {
				internalErr(w, err)
				return
			}
			defer rows.Close()
			out := []SessionSet{}
			for rows.Next() {
				var s SessionSet
				if err := rows.Scan(&s.ID, &s.SessionID, &s.ExercicioID, &s.Series, &s.Repeticoes, &s.CargaKg, &s.RIR, &s.Completed, &s.Notes); err != nil {
					internalErr(w, err)
					return
				}
				out = append(out, s)
			}
			jsonWrite(w, http.StatusOK, map[string]any{"items": out})
			return

		case http.MethodPost:
			var in SessionSet
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				badRequest(w, "invalid json")
				return
			}
			in.SessionID = sessionID
			err := db.QueryRow(`
			   INSERT INTO workout_sets (session_id, exercicio_id, series, repeticoes, carga_kg, rir, completed, notes)
			   VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
				in.SessionID, in.ExercicioID, in.Series, in.Repeticoes, in.CargaKg, in.RIR, in.Completed, in.Notes,
			).Scan(&in.ID)
			if err != nil {
				internalErr(w, err)
				return
			}
			jsonWrite(w, http.StatusCreated, in)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// PATCH /api/sets/{id}
// DELETE /api/sets/{id}
func SetItem(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		// /api /sets /{id}
		if len(parts) < 3 {
			badRequest(w, "missing set id")
			return
		}
		setID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || setID <= 0 {
			badRequest(w, "invalid set id")
			return
		}
		switch r.Method {
		case http.MethodPatch:
			var in SessionSet
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				badRequest(w, "invalid json")
				return
			}
			// atualiza apenas campos enviados
			q := `UPDATE workout_sets SET `
			args := []any{}
			i := 1
			if in.ExercicioID != 0 {
				q += `exercicio_id = $` + itoa(i) + `, `
				args = append(args, in.ExercicioID)
				i++
			}
			if in.Series != 0 {
				q += `series = $` + itoa(i) + `, `
				args = append(args, in.Series)
				i++
			}
			if in.Repeticoes != 0 {
				q += `repeticoes = $` + itoa(i) + `, `
				args = append(args, in.Repeticoes)
				i++
			}
			q += `carga_kg = COALESCE($` + itoa(i) + `, carga_kg), `
			args = append(args, nullFloat(in.CargaKg))
			i++
			q += `rir = COALESCE($` + itoa(i) + `, rir), `
			args = append(args, nullInt(in.RIR))
			i++
			q += `completed = COALESCE($` + itoa(i) + `, completed), `
			args = append(args, nullBoolPtr(in.Completed))
			i++
			q += `notes = COALESCE($` + itoa(i) + `, notes) `
			args = append(args, nullString(in.Notes))
			i++
			q += `WHERE id = $` + itoa(i)
			args = append(args, setID)
			if _, err := db.Exec(q, args...); err != nil {
				internalErr(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return

		case http.MethodDelete:
			if _, err := db.Exec(`DELETE FROM workout_sets WHERE id=$1`, setID); err != nil {
				internalErr(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

// utils nullables
func nullFloat(v float64) any {
	if v == 0 {
		return nil
	}
	return v
}
func nullInt(v int) any {
	if v == 0 {
		return nil
	}
	return v
}
func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func nullBoolPtr(b bool) any {
	// false=>nil mantém simples no PATCH parcial
	if !b {
		return nil
	}
	return true
}

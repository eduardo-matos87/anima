package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type updateSessionReq struct {
	SessionAt *time.Time `json:"session_at,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

// PATCH /api/sessions/update/{id}
// DELETE /api/sessions/update/{id}
func UpdateDeleteSession(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			badRequest(w, "missing id")
			return
		}
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || id <= 0 {
			badRequest(w, "invalid id")
			return
		}

		switch r.Method {
		case http.MethodDelete:
			if _, err := db.Exec(`DELETE FROM workout_sessions WHERE id=$1`, id); err != nil {
				internalErr(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return

		case http.MethodPatch:
			var in updateSessionReq
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				badRequest(w, "invalid json")
				return
			}
			if in.SessionAt == nil && in.Notes == nil {
				badRequest(w, "no fields to update")
				return
			}
			q := `UPDATE workout_sessions SET `
			args := []any{}
			i := 1
			if in.SessionAt != nil {
				q += `session_at = $` + itoa(i)
				args = append(args, *in.SessionAt)
				i++
			}
			if in.Notes != nil {
				if len(args) > 0 {
					q += `, `
				}
				q += `notes = $` + itoa(i)
				args = append(args, *in.Notes)
				i++
			}
			q += `, updated_at = NOW() WHERE id = $` + itoa(i)
			args = append(args, id)
			if _, err := db.Exec(q, args...); err != nil {
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

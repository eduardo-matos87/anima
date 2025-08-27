package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// PATCH /api/sessions/update/{id}
// Campos permitidos: completed(bool), duration_min(int), rpe_session(int 1..10), notes(string), session_at(RFC3339)
func SessionsPatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(GetUserID(r))

	// Extrai ID do fim do path (funciona para /api/sessions/update/{id} e /api/sessions/{id})
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		badRequest(w, "invalid path")
		return
	}
	// .../sessions/{id} OU .../sessions/update/{id}
	idPart := parts[len(parts)-1]
	id, ok := toInt64(idPart)
	if !ok || id <= 0 {
		badRequest(w, "invalid session id")
		return
	}

	var in map[string]any
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if len(in) == 0 {
		badRequest(w, "empty body")
		return
	}

	setParts := []string{}
	args := []any{}
	argIdx := 1

	for k, v := range in {
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

		case "duration_min":
			iv, ok := toInt64(v)
			if !ok || iv < 0 || iv > 300 {
				badRequest(w, "duration_min must be 0..300")
				return
			}
			setParts = append(setParts, "duration_min = $"+itoa(argIdx))
			args = append(args, iv)
			argIdx++

		case "rpe_session":
			iv, ok := toInt64(v)
			if !ok || iv < 1 || iv > 10 {
				badRequest(w, "rpe_session must be 1..10")
				return
			}
			setParts = append(setParts, "rpe_session = $"+itoa(argIdx))
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

		case "session_at":
			sv, ok := v.(string)
			if !ok || strings.TrimSpace(sv) == "" {
				badRequest(w, "session_at must be RFC3339 string")
				return
			}
			t, err := time.Parse(time.RFC3339, sv)
			if err != nil {
				badRequest(w, "invalid session_at (RFC3339)")
				return
			}
			setParts = append(setParts, "session_at = $"+itoa(argIdx))
			args = append(args, t)
			argIdx++

		default:
			badRequest(w, "unsupported field: "+k)
			return
		}
	}

	if len(setParts) == 0 {
		badRequest(w, "no updatable fields")
		return
	}

	q := `
UPDATE workout_sessions
SET ` + strings.Join(setParts, ", ") + `, updated_at = NOW()
WHERE id = $` + itoa(argIdx) + `
  AND ($` + itoa(argIdx+1) + ` = '' OR user_id IS NULL OR user_id = $` + itoa(argIdx+1) + `)
RETURNING id, treino_id, session_at, completed, duration_min, rpe_session, notes, user_id, updated_at
`
	args = append(args, id, userID)

	var out struct {
		ID          int64     `json:"id"`
		TreinoID    int64     `json:"treino_id"`
		SessionAt   time.Time `json:"session_at"`
		Completed   bool      `json:"completed"`
		DurationMin *int64    `json:"duration_min,omitempty"`
		RPESession  *int64    `json:"rpe_session,omitempty"`
		Notes       *string   `json:"notes,omitempty"`
		UserID      *string   `json:"user_id,omitempty"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	var (
		duration sql.NullInt64
		rpe      sql.NullInt64
		notes    sql.NullString
		uid      sql.NullString
	)
	err := sessionsDB.QueryRow(q, args...).Scan(
		&out.ID, &out.TreinoID, &out.SessionAt, &out.Completed, &duration, &rpe, &notes, &uid, &out.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// não existe ou não pertence ao user
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalErr(w, err)
		return
	}
	if duration.Valid {
		out.DurationMin = &duration.Int64
	}
	if rpe.Valid {
		out.RPESession = &rpe.Int64
	}
	if notes.Valid {
		s := notes.String
		out.Notes = &s
	}
	if uid.Valid {
		s := uid.String
		out.UserID = &s
	}

	jsonWrite(w, http.StatusOK, out)
}

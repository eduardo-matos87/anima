package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func SessionsCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}

	// Dono da sessão (JWT ou header X-User-ID)
	userID := strings.TrimSpace(GetUserID(r))

	var in struct {
		TreinoID  int64   `json:"treino_id"`
		SessionAt *string `json:"session_at,omitempty"` // RFC3339 opcional
		Notes     *string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if in.TreinoID <= 0 {
		badRequest(w, "treino_id required")
		return
	}

	// session_at: agora (UTC) se não vier
	sessionAt := time.Now().UTC()
	if in.SessionAt != nil && strings.TrimSpace(*in.SessionAt) != "" {
		t, err := time.Parse(time.RFC3339, *in.SessionAt)
		if err != nil {
			badRequest(w, "invalid session_at (RFC3339)")
			return
		}
		sessionAt = t
	}

	notes := ""
	if in.Notes != nil {
		notes = *in.Notes
	}

	const q = `
INSERT INTO workout_sessions (treino_id, session_at, notes, user_id)
VALUES ($1, $2, $3, NULLIF($4, ''))
RETURNING id
`
	var id int64
	if err := sessionsDB.QueryRow(q, in.TreinoID, sessionAt, notes, userID).Scan(&id); err != nil {
		internalErr(w, err)
		return
	}
	jsonWrite(w, http.StatusCreated, map[string]any{"id": id})
}

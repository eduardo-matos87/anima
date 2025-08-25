package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type SessionCreateReq struct {
	TreinoID  *int64     `json:"treino_id,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

func SessionsCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		badRequest(w, "use POST")
		return
	}
	uid := getUserID(r)

	var req SessionCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid json")
		return
	}

	start := time.Now().UTC()
	if req.StartedAt != nil {
		start = req.StartedAt.UTC()
	}

	tx, err := sessionsDB.Begin()
	if err != nil {
		badRequest(w, "db error")
		return
	}
	defer tx.Rollback()

	const q = `
INSERT INTO workout_sessions (user_id, treino_id, started_at, notes)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at`
	var id int64
	var created time.Time
	err = tx.QueryRow(q, uid, req.TreinoID, start, req.Notes).Scan(&id, &created)
	if err != nil {
		badRequest(w, "insert error")
		return
	}
	if err = tx.Commit(); err != nil {
		badRequest(w, "commit error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id": id,
	})
}

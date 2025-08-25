package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SessionPatchReq struct {
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	DurationSec *int64     `json:"duration_sec,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
}

func SessionsPatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		badRequest(w, "use PATCH")
		return
	}
	uid := getUserID(r)

	// /api/sessions/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		badRequest(w, "missing id")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		badRequest(w, "invalid id")
		return
	}

	var req SessionPatchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid json")
		return
	}

	// Monta SET dinâmico simples
	setCols := []string{}
	args := []any{}
	argi := 1
	if req.EndedAt != nil {
		setCols = append(setCols, "ended_at = $"+strconv.Itoa(argi))
		args = append(args, req.EndedAt.UTC())
		argi++
	}
	if req.DurationSec != nil {
		setCols = append(setCols, "duration_sec = $"+strconv.Itoa(argi))
		args = append(args, *req.DurationSec)
		argi++
	}
	if req.Notes != nil {
		setCols = append(setCols, "notes = $"+strconv.Itoa(argi))
		args = append(args, *req.Notes)
		argi++
	}
	if len(setCols) == 0 {
		badRequest(w, "no fields to update")
		return
	}

	q := "UPDATE workout_sessions SET " + strings.Join(setCols, ", ") + " WHERE id = $" + strconv.Itoa(argi) + " AND user_id = $" + strconv.Itoa(argi+1)
	args = append(args, id, uid)
	res, err := sessionsDB.Exec(q, args...)
	if err != nil {
		badRequest(w, "db error")
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		notFound(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func SessionsDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		badRequest(w, "use DELETE")
		return
	}
	uid := getUserID(r)

	// /api/sessions/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		badRequest(w, "missing id")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		badRequest(w, "invalid id")
		return
	}

	const q = `DELETE FROM workout_sessions WHERE id = $1 AND user_id = $2`
	res, err := sessionsDB.Exec(q, id, uid)
	if err != nil {
		badRequest(w, "db error")
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		notFound(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

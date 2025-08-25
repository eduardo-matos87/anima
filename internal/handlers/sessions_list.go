package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
)

func SessionsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		badRequest(w, "use GET")
		return
	}
	uid := getUserID(r)
	q := r.URL.Query()

	limit := 20
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	const qCount = `SELECT COUNT(*) FROM workout_sessions WHERE user_id = $1`
	var count int64
	if err := sessionsDB.QueryRow(qCount, uid).Scan(&count); err != nil {
		badRequest(w, "db error")
		return
	}

	const qList = `
SELECT id, user_id, treino_id, started_at, ended_at, duration_sec, notes, created_at
FROM workout_sessions
WHERE user_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3`
	rows, err := sessionsDB.Query(qList, uid, limit, offset)
	if err != nil {
		badRequest(w, "db error")
		return
	}
	defer rows.Close()

	items := make([]Session, 0, limit)
	for rows.Next() {
		var s Session
		var treinoID sql.NullInt64
		var ended sql.NullTime
		var dur sql.NullInt64
		var notes sql.NullString
		if err := rows.Scan(&s.ID, &s.UserID, &treinoID, &s.StartedAt, &ended, &dur, &notes, &s.CreatedAt); err != nil {
			badRequest(w, "scan error")
			return
		}
		if treinoID.Valid {
			v := treinoID.Int64
			s.TreinoID = &v
		}
		if ended.Valid {
			v := ended.Time
			s.EndedAt = &v
		}
		if dur.Valid {
			v := dur.Int64
			s.DurationSec = &v
		}
		if notes.Valid {
			v := notes.String
			s.Notes = &v
		}
		items = append(items, s)
	}

	var next *int
	if int64(offset+limit) < count {
		n := offset + limit
		next = &n
	}
	writeJSON(w, http.StatusOK, Page[Session]{Items: items, Next: next, Count: count})
}

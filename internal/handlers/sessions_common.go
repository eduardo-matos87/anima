package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

var sessionsDB *sql.DB

func SetSessionsDB(db *sql.DB) { sessionsDB = db }

// (REMOVIDO) getUserID — usar a versão já existente em userctx.go

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

type Session struct {
	ID          int64      `json:"id"`
	UserID      string     `json:"user_id"`
	TreinoID    *int64     `json:"treino_id,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	DurationSec *int64     `json:"duration_sec,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type WorkoutSet struct {
	ID          int64     `json:"id"`
	SessionID   int64     `json:"session_id"`
	ExercicioID int64     `json:"exercicio_id"`
	SetIndex    int       `json:"set_index"`
	WeightKG    *float64  `json:"weight_kg,omitempty"`
	Reps        *int      `json:"reps,omitempty"`
	RIR         *int      `json:"rir,omitempty"`
	Completed   bool      `json:"completed"`
	RestSec     *int      `json:"rest_sec,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Page[T any] struct {
	Items []T   `json:"items"`
	Next  *int  `json:"next,omitempty"`
	Count int64 `json:"count"`
}

func queryOneSession(id int64, userID string) (*Session, error) {
	const q = `
SELECT id, user_id, treino_id, started_at, ended_at, duration_sec, notes, created_at
FROM workout_sessions
WHERE id = $1 AND user_id = $2`
	row := sessionsDB.QueryRow(q, id, userID)
	var s Session
	var treinoID sql.NullInt64
	var endedAt sql.NullTime
	var duration sql.NullInt64
	var notes sql.NullString
	err := row.Scan(&s.ID, &s.UserID, &treinoID, &s.StartedAt, &endedAt, &duration, &notes, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	if treinoID.Valid {
		val := treinoID.Int64
		s.TreinoID = &val
	}
	if endedAt.Valid {
		val := endedAt.Time
		s.EndedAt = &val
	}
	if duration.Valid {
		val := duration.Int64
		s.DurationSec = &val
	}
	if notes.Valid {
		val := notes.String
		s.Notes = &val
	}
	return &s, nil
}

func errIsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

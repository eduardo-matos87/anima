package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Session struct {
	ID        int64     `json:"id"`
	TreinoID  int64     `json:"treino_id"`
	SessionAt time.Time `json:"session_at"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SessionSet struct {
	ID          int64   `json:"id"`
	SessionID   int64   `json:"session_id"`
	ExercicioID int64   `json:"exercicio_id"`
	Series      int     `json:"series"`
	Repeticoes  int     `json:"repeticoes"`
	CargaKg     float64 `json:"carga_kg,omitempty"`
	RIR         int     `json:"rir,omitempty"`
	Completed   bool    `json:"completed"`
	Notes       string  `json:"notes,omitempty"`
}

func parseIntQuery(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func parseInt64Query(r *http.Request, key string, def int64) int64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return i
}

func parseTimeQuery(r *http.Request, key string) (time.Time, bool) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return time.Time{}, false
	}
	// aceita RFC3339 ou yyyy-mm-dd
	t, err := time.Parse(time.RFC3339, v)
	if err == nil {
		return t, true
	}
	t, err = time.Parse("2006-01-02", v)
	return t, err == nil
}

func jsonWrite(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	jsonWrite(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func notFound(w http.ResponseWriter) {
	jsonWrite(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func internalErr(w http.ResponseWriter, err error) {
	jsonWrite(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

// Helpers de DB
func fetchSession(ctx context.Context, db *sql.DB, id int64) (*Session, error) {
	const q = `
	  SELECT id, treino_id, session_at, COALESCE(notes,''), created_at, updated_at
	  FROM workout_sessions WHERE id = $1`
	var s Session
	err := db.QueryRowContext(ctx, q, id).Scan(&s.ID, &s.TreinoID, &s.SessionAt, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func fetchSet(ctx context.Context, db *sql.DB, id int64) (*SessionSet, error) {
	const q = `
	  SELECT id, session_id, exercicio_id, series, repeticoes, COALESCE(carga_kg,0), COALESCE(rir,0), completed, COALESCE(notes,'')
	  FROM workout_sets WHERE id = $1`
	var s SessionSet
	err := db.QueryRowContext(ctx, q, id).Scan(
		&s.ID, &s.SessionID, &s.ExercicioID, &s.Series, &s.Repeticoes, &s.CargaKg, &s.RIR, &s.Completed, &s.Notes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Validação simples
func validateSessionPayload(treinoID int64, at time.Time) error {
	if treinoID <= 0 {
		return fmt.Errorf("treino_id inválido")
	}
	if at.IsZero() {
		return fmt.Errorf("session_at inválido")
	}
	return nil
}

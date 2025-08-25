package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type SetCreateReq struct {
	ExercicioID int64    `json:"exercicio_id"`
	SetIndex    int      `json:"set_index"`
	WeightKG    *float64 `json:"weight_kg,omitempty"`
	Reps        *int     `json:"reps,omitempty"`
	RIR         *int     `json:"rir,omitempty"`
	Completed   *bool    `json:"completed,omitempty"`
	RestSec     *int     `json:"rest_sec,omitempty"`
}

func SetsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		badRequest(w, "use GET")
		return
	}
	uid := getUserID(r)
	// /api/sessions/{id}/sets
	p := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(p, "/")
	if len(parts) < 2 || parts[1] != "sets" {
		badRequest(w, "invalid path")
		return
	}
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		badRequest(w, "invalid session id")
		return
	}

	// garantir ownership da sessão
	if _, err := queryOneSession(sessionID, uid); err != nil {
		if errIsNotFound(err) {
			notFound(w)
			return
		}
		badRequest(w, "db error")
		return
	}

	const q = `
SELECT id, session_id, exercicio_id, set_index, weight_kg, reps, rir, completed, rest_sec, created_at
FROM workout_sets
WHERE session_id = $1
ORDER BY exercicio_id, set_index`
	rows, err := sessionsDB.Query(q, sessionID)
	if err != nil {
		badRequest(w, "db error")
		return
	}
	defer rows.Close()

	var items []WorkoutSet
	for rows.Next() {
		var ws WorkoutSet
		var weight sql.NullFloat64
		var reps, rir, rest sql.NullInt64
		if err := rows.Scan(&ws.ID, &ws.SessionID, &ws.ExercicioID, &ws.SetIndex, &weight, &reps, &rir, &ws.Completed, &rest, &ws.CreatedAt); err != nil {
			badRequest(w, "scan error")
			return
		}
		if weight.Valid {
			v := weight.Float64
			ws.WeightKG = &v
		}
		if reps.Valid {
			v := int(reps.Int64)
			ws.Reps = &v
		}
		if rir.Valid {
			v := int(rir.Int64)
			ws.RIR = &v
		}
		if rest.Valid {
			v := int(rest.Int64)
			ws.RestSec = &v
		}
		items = append(items, ws)
	}
	writeJSON(w, http.StatusOK, items)
}

func SetsCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		badRequest(w, "use POST")
		return
	}
	uid := getUserID(r)
	// /api/sessions/{id}/sets
	p := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(p, "/")
	if len(parts) < 2 || parts[1] != "sets" {
		badRequest(w, "invalid path")
		return
	}
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		badRequest(w, "invalid session id")
		return
	}
	// ownership
	if _, err := queryOneSession(sessionID, uid); err != nil {
		if errIsNotFound(err) {
			notFound(w)
			return
		}
		badRequest(w, "db error")
		return
	}

	var req SetCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid json")
		return
	}
	completed := true
	if req.Completed != nil {
		completed = *req.Completed
	}

	const q = `
INSERT INTO workout_sets (session_id, exercicio_id, set_index, weight_kg, reps, rir, completed, rest_sec)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING id`
	var id int64
	err = sessionsDB.QueryRow(q, sessionID, req.ExercicioID, req.SetIndex, req.WeightKG, req.Reps, req.RIR, completed, req.RestSec).Scan(&id)
	if err != nil {
		badRequest(w, "insert error (set may be duplicate set_index?)")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

type SetPatchReq struct {
	WeightKG  *float64 `json:"weight_kg,omitempty"`
	Reps      *int     `json:"reps,omitempty"`
	RIR       *int     `json:"rir,omitempty"`
	Completed *bool    `json:"completed,omitempty"`
	RestSec   *int     `json:"rest_sec,omitempty"`
	SetIndex  *int     `json:"set_index,omitempty"`
}

func SetsPatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		badRequest(w, "use PATCH")
		return
	}
	uid := getUserID(r)
	// /api/sessions/{id}/sets/{set_id}
	p := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(p, "/")
	if len(parts) < 3 || parts[1] != "sets" {
		badRequest(w, "invalid path")
		return
	}
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		badRequest(w, "invalid session id")
		return
	}
	setID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		badRequest(w, "invalid set id")
		return
	}

	// ownership
	if _, err := queryOneSession(sessionID, uid); err != nil {
		if errIsNotFound(err) {
			notFound(w)
			return
		}
		badRequest(w, "db error")
		return
	}

	var req SetPatchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid json")
		return
	}

	cols := []string{}
	args := []any{}
	i := 1
	if req.WeightKG != nil {
		cols = append(cols, "weight_kg = $"+strconv.Itoa(i))
		args = append(args, *req.WeightKG)
		i++
	}
	if req.Reps != nil {
		cols = append(cols, "reps = $"+strconv.Itoa(i))
		args = append(args, *req.Reps)
		i++
	}
	if req.RIR != nil {
		cols = append(cols, "rir = $"+strconv.Itoa(i))
		args = append(args, *req.RIR)
		i++
	}
	if req.Completed != nil {
		cols = append(cols, "completed = $"+strconv.Itoa(i))
		args = append(args, *req.Completed)
		i++
	}
	if req.RestSec != nil {
		cols = append(cols, "rest_sec = $"+strconv.Itoa(i))
		args = append(args, *req.RestSec)
		i++
	}
	if req.SetIndex != nil {
		cols = append(cols, "set_index = $"+strconv.Itoa(i))
		args = append(args, *req.SetIndex)
		i++
	}

	if len(cols) == 0 {
		badRequest(w, "no fields to update")
		return
	}

	q := "UPDATE workout_sets SET " + strings.Join(cols, ", ") + " WHERE id = $" + strconv.Itoa(i) + " AND session_id = $" + strconv.Itoa(i+1)
	args = append(args, setID, sessionID)

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

func SetsDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		badRequest(w, "use DELETE")
		return
	}
	uid := getUserID(r)
	// /api/sessions/{id}/sets/{set_id}
	p := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(p, "/")
	if len(parts) < 3 || parts[1] != "sets" {
		badRequest(w, "invalid path")
		return
	}
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		badRequest(w, "invalid session id")
		return
	}
	setID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		badRequest(w, "invalid set id")
		return
	}

	// ownership
	if _, err := queryOneSession(sessionID, uid); err != nil {
		if errIsNotFound(err) {
			notFound(w)
			return
		}
		badRequest(w, "db error")
		return
	}

	const q = `DELETE FROM workout_sets WHERE id = $1 AND session_id = $2`
	res, err := sessionsDB.Exec(q, setID, sessionID)
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

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Fonte única do DB neste package.
var sessionsDB *sql.DB

// Injeção feita pelo main.go
func SetSessionsDB(db *sql.DB) { sessionsDB = db }

// =======================
// Sessions (leitura segura)
// =======================

// GET /api/sessions
// Query params: page, page_size, treino_id, from, to
// Guarda por dono: ($user=” OR ws.user_id = $user)
func SessionsList(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(GetUserID(r))

	// paginação
	page := int64(1)
	pageSize := int64(20)
	if v := r.URL.Query().Get("page"); v != "" {
		if n, ok := toInt64(v); ok && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		if n, ok := toInt64(v); ok && n > 0 {
			pageSize = n
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// filtros
	var (
		whereParts []string
		args       []any
		argIdx     = 1
	)
	// guarda de dono
	whereParts = append(whereParts, "($"+itoa(argIdx)+" = '' OR ws.user_id = $"+itoa(argIdx)+")")
	args = append(args, userID)
	argIdx++

	// treino_id
	if v := r.URL.Query().Get("treino_id"); v != "" {
		if n, ok := toInt64(v); ok && n > 0 {
			whereParts = append(whereParts, "ws.treino_id = $"+itoa(argIdx))
			args = append(args, n)
			argIdx++
		} else {
			badRequest(w, "invalid treino_id")
			return
		}
	}

	// from/to (RFC3339) em started_at
	parseTS := func(s string) (time.Time, bool) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, ok := parseTS(v); ok {
			whereParts = append(whereParts, "ws.started_at >= $"+itoa(argIdx))
			args = append(args, t)
			argIdx++
		} else {
			badRequest(w, "invalid from (RFC3339)")
			return
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, ok := parseTS(v); ok {
			whereParts = append(whereParts, "ws.started_at <= $"+itoa(argIdx))
			args = append(args, t)
			argIdx++
		} else {
			badRequest(w, "invalid to (RFC3339)")
			return
		}
	}

	whereSQL := "WHERE " + strings.Join(whereParts, " AND ")

	// total
	countQ := `SELECT COUNT(*) FROM workout_sessions ws ` + whereSQL
	var total int64
	if err := sessionsDB.QueryRow(countQ, args...).Scan(&total); err != nil {
		internalErr(w, err)
		return
	}

	// page data
	listQ := `
SELECT ws.id, ws.treino_id, ws.started_at, COALESCE(ws.notes,''),
       ws.user_id, ws.ended_at, ws.duration_sec, ws.created_at
FROM workout_sessions ws
` + whereSQL + `
ORDER BY ws.started_at DESC, ws.id DESC
LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := sessionsDB.Query(listQ, args...)
	if err != nil {
		internalErr(w, err)
		return
	}
	defer rows.Close()

	type rowOut struct {
		ID          int64      `json:"id"`
		TreinoID    *int64     `json:"treino_id,omitempty"`
		SessionAt   time.Time  `json:"session_at"` // mapeia started_at
		Notes       string     `json:"notes"`
		UserID      string     `json:"user_id"`
		EndedAt     *time.Time `json:"ended_at,omitempty"`
		DurationSec *int64     `json:"duration_sec,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	items := make([]rowOut, 0, pageSize)
	for rows.Next() {
		var (
			r   rowOut
			tid sql.NullInt64
			en  sql.NullTime
			ds  sql.NullInt64
		)
		if err := rows.Scan(
			&r.ID, &tid, &r.SessionAt, &r.Notes,
			&r.UserID, &en, &ds, &r.CreatedAt,
		); err != nil {
			internalErr(w, err)
			return
		}
		if tid.Valid {
			v := tid.Int64
			r.TreinoID = &v
		}
		if en.Valid {
			t := en.Time
			r.EndedAt = &t
		}
		if ds.Valid {
			v := ds.Int64
			r.DurationSec = &v
		}
		items = append(items, r)
	}

	jsonWrite(w, http.StatusOK, map[string]any{
		"items":      items,
		"page":       page,
		"page_size":  pageSize,
		"total_hint": total,
	})
}

// GET /api/sessions/{id}
func SessionsGet(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(GetUserID(r))

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		badRequest(w, "invalid path")
		return
	}
	// .../sessions/{id}
	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || id <= 0 {
		badRequest(w, "invalid session id")
		return
	}

	q := `
SELECT ws.id, ws.treino_id, ws.started_at, COALESCE(ws.notes,''),
       ws.user_id, ws.ended_at, ws.duration_sec, ws.created_at
FROM workout_sessions ws
WHERE ws.id = $1
  AND ($2 = '' OR ws.user_id = $2)
`
	var (
		out struct {
			ID          int64      `json:"id"`
			TreinoID    *int64     `json:"treino_id,omitempty"`
			SessionAt   time.Time  `json:"session_at"`
			Notes       string     `json:"notes"`
			UserID      string     `json:"user_id"`
			EndedAt     *time.Time `json:"ended_at,omitempty"`
			DurationSec *int64     `json:"duration_sec,omitempty"`
			CreatedAt   time.Time  `json:"created_at"`
		}
		tid sql.NullInt64
		en  sql.NullTime
		ds  sql.NullInt64
	)
	err = sessionsDB.QueryRow(q, id, userID).Scan(
		&out.ID, &tid, &out.SessionAt, &out.Notes,
		&out.UserID, &en, &ds, &out.CreatedAt,
	)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalErr(w, err)
		return
	}
	if tid.Valid {
		v := tid.Int64
		out.TreinoID = &v
	}
	if en.Valid {
		t := en.Time
		out.EndedAt = &t
	}
	if ds.Valid {
		v := ds.Int64
		out.DurationSec = &v
	}
	jsonWrite(w, http.StatusOK, out)
}

// POST /api/sessions (factory precisa de *sql.DB)
func SessionsCreate(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := strings.TrimSpace(GetUserID(r))
		if userID == "" {
			badRequest(w, "missing X-User-ID")
			return
		}

		var in struct {
			TreinoID  *int64  `json:"treino_id,omitempty"`
			Notes     *string `json:"notes,omitempty"`
			SessionAt *string `json:"session_at,omitempty"` // compat; mapeia started_at
			StartedAt *string `json:"started_at,omitempty"` // preferível
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			badRequest(w, "invalid json")
			return
		}

		var startedAt time.Time
		switch {
		case in.StartedAt != nil && *in.StartedAt != "":
			t, err := time.Parse(time.RFC3339, *in.StartedAt)
			if err != nil {
				badRequest(w, "started_at must be RFC3339")
				return
			}
			startedAt = t
		case in.SessionAt != nil && *in.SessionAt != "":
			t, err := time.Parse(time.RFC3339, *in.SessionAt)
			if err != nil {
				badRequest(w, "session_at must be RFC3339")
				return
			}
			startedAt = t
		default:
			startedAt = time.Now()
		}

		var id int64
		err := db.QueryRow(`
INSERT INTO workout_sessions (user_id, treino_id, started_at, notes)
VALUES ($1, $2, $3, $4)
RETURNING id
`, userID, in.TreinoID, startedAt, in.Notes).Scan(&id)
		if err != nil {
			internalErr(w, err)
			return
		}
		jsonWrite(w, http.StatusCreated, map[string]any{"id": id})
	})
}

// PATCH / DELETE /api/sessions/update/{id}
func SessionsPatch(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(GetUserID(r))

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/update/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		badRequest(w, "invalid path")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
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

	allowed := map[string]bool{
		"notes":        true,
		"treino_id":    true,
		"started_at":   true,
		"session_at":   true, // compat
		"ended_at":     true,
		"duration_sec": true,
	}

	setParts := []string{}
	args := []any{}
	argIdx := 1

	for k, v := range in {
		if !allowed[k] {
			badRequest(w, "unsupported field: "+k)
			return
		}
		switch k {
		case "notes":
			s, ok := v.(string)
			if !ok {
				badRequest(w, "notes must be string")
				return
			}
			setParts = append(setParts, "notes = $"+itoa(argIdx))
			args = append(args, s)
			argIdx++
		case "treino_id":
			if v == nil {
				setParts = append(setParts, "treino_id = NULL")
				continue
			}
			iv, ok := toInt64(v)
			if !ok {
				badRequest(w, "treino_id must be int or null")
				return
			}
			setParts = append(setParts, "treino_id = $"+itoa(argIdx))
			args = append(args, iv)
			argIdx++
		case "started_at", "session_at":
			if v == nil {
				badRequest(w, k+" must be RFC3339")
				return
			}
			s, ok := v.(string)
			if !ok {
				badRequest(w, k+" must be RFC3339")
				return
			}
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				badRequest(w, k+" must be RFC3339")
				return
			}
			setParts = append(setParts, "started_at = $"+itoa(argIdx))
			args = append(args, t)
			argIdx++
		case "ended_at":
			if v == nil {
				setParts = append(setParts, "ended_at = NULL")
				continue
			}
			s, ok := v.(string)
			if !ok {
				badRequest(w, "ended_at must be RFC3339 or null")
				return
			}
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				badRequest(w, "ended_at must be RFC3339 or null")
				return
			}
			setParts = append(setParts, "ended_at = $"+itoa(argIdx))
			args = append(args, t)
			argIdx++
		case "duration_sec":
			if v == nil {
				setParts = append(setParts, "duration_sec = NULL")
				continue
			}
			iv, ok := toInt64(v)
			if !ok {
				badRequest(w, "duration_sec must be int or null")
				return
			}
			setParts = append(setParts, "duration_sec = $"+itoa(argIdx))
			args = append(args, iv)
			argIdx++
		}
	}

	q := `
UPDATE workout_sessions
SET ` + strings.Join(setParts, ", ") + `
WHERE id = $` + itoa(argIdx) + ` AND ($` + itoa(argIdx+1) + ` = '' OR user_id = $` + itoa(argIdx+1) + `)
RETURNING id
`
	args = append(args, id, userID)
	var rid int64
	if err := sessionsDB.QueryRow(q, args...).Scan(&rid); err != nil {
		http.NotFound(w, r)
		return
	}
	jsonWrite(w, http.StatusOK, map[string]any{"id": rid})
}

func SessionsDelete(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := strings.TrimSpace(GetUserID(r))

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/update/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		badRequest(w, "invalid path")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		badRequest(w, "invalid session id")
		return
	}

	res, err := sessionsDB.Exec(`
DELETE FROM workout_sessions
WHERE id = $1 AND ($2 = '' OR user_id = $2)
`, id, userID)
	if err != nil {
		internalErr(w, err)
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============
// Overload GET (compat)
// ============

func NextLoad(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	OverloadSuggest(sessionsDB).ServeHTTP(w, r)
}

// ==========================
// Sets compat com seu main:
//  - GET/POST     /api/sessions/{id}/sets
//  - PATCH/DELETE /api/sets/{id}
// (usa schema novo: set_index/weight_kg/reps/rir/completed/rest_sec)
// Aceita também JSON antigo: series/repeticoes/carga_kg
// ==========================

func SetsList(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	// path: /api/sessions/{id}/sets
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) < 2 || parts[1] != "sets" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || sessionID <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid := GetUserID(r)
	// ownership
	if uid != "" {
		var ok bool
		if err := sessionsDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM workout_sessions WHERE id=$1 AND user_id=$2)`, sessionID, uid).Scan(&ok); err != nil || !ok {
			http.NotFound(w, r)
			return
		}
	}

	rows, err := sessionsDB.Query(`
		SELECT id, session_id, exercicio_id, set_index,
		       weight_kg, reps, rir, completed, COALESCE(rest_sec,0), created_at
		FROM workout_sets
		WHERE session_id = $1
		ORDER BY set_index ASC, id ASC
	`, sessionID)
	if err != nil {
		internalErr(w, err)
		return
	}
	defer rows.Close()

	type item struct {
		ID        int64    `json:"id"`
		SessionID int64    `json:"session_id"`
		Exercicio int      `json:"exercicio_id"`
		SetIndex  int      `json:"set_index"`
		WeightKg  *float64 `json:"weight_kg,omitempty"`
		Reps      *int     `json:"reps,omitempty"`
		RIR       *int     `json:"rir,omitempty"`
		Completed bool     `json:"completed"`
		RestSec   *int     `json:"rest_sec,omitempty"`
		CreatedAt string   `json:"created_at"`
	}
	var items []item
	for rows.Next() {
		var it item
		var wkg sql.NullFloat64
		var reps sql.NullInt64
		var rir sql.NullInt64
		var rest sql.NullInt64
		if err := rows.Scan(&it.ID, &it.SessionID, &it.Exercicio, &it.SetIndex,
			&wkg, &reps, &rir, &it.Completed, &rest, &it.CreatedAt); err != nil {
			internalErr(w, err)
			return
		}
		if wkg.Valid {
			v := wkg.Float64
			it.WeightKg = &v
		}
		if reps.Valid {
			v := int(reps.Int64)
			it.Reps = &v
		}
		if rir.Valid {
			v := int(rir.Int64)
			it.RIR = &v
		}
		if rest.Valid {
			v := int(rest.Int64)
			it.RestSec = &v
		}
		items = append(items, it)
	}
	jsonWrite(w, http.StatusOK, map[string]any{"items": items})
}

func SetsCreate(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	// path: /api/sessions/{id}/sets
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) < 2 || parts[1] != "sets" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || sessionID <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	uid := GetUserID(r)
	if uid != "" {
		var ok bool
		if err := sessionsDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM workout_sessions WHERE id=$1 AND user_id=$2)`, sessionID, uid).Scan(&ok); err != nil || !ok {
			http.NotFound(w, r)
			return
		}
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		badRequest(w, "invalid json")
		return
	}

	getInt := func(keys ...string) (*int, bool) {
		for _, k := range keys {
			if v, ok := body[k]; ok && v != nil {
				switch t := v.(type) {
				case float64:
					iv := int(t)
					return &iv, true
				case int:
					iv := t
					return &iv, true
				}
			}
		}
		return nil, false
	}
	getFloat := func(keys ...string) (*float64, bool) {
		for _, k := range keys {
			if v, ok := body[k]; ok && v != nil {
				switch t := v.(type) {
				case float64:
					fv := t
					return &fv, true
				case int:
					fv := float64(t)
					return &fv, true
				}
			}
		}
		return nil, false
	}

	exID, ok := getInt("exercicio_id")
	if !ok || exID == nil || *exID <= 0 {
		badRequest(w, "exercicio_id required")
		return
	}

	setIdx, ok := getInt("set_index", "series")
	if !ok || setIdx == nil || *setIdx <= 0 {
		var next int
		_ = sessionsDB.QueryRow(`SELECT COALESCE(MAX(set_index),0)+1 FROM workout_sets WHERE session_id=$1`, sessionID).Scan(&next)
		setIdx = &next
	}

	weight, _ := getFloat("weight_kg", "carga_kg")
	reps, _ := getInt("reps", "repeticoes")
	rir, _ := getInt("rir")
	rest, _ := getInt("rest_sec")

	completed := false
	if v, ok := body["completed"]; ok {
		if b, ok := v.(bool); ok {
			completed = b
		}
	}

	var id int64
	err = sessionsDB.QueryRow(`
		INSERT INTO workout_sets
		  (session_id, exercicio_id, set_index, weight_kg, reps, rir, completed, rest_sec)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
	`, sessionID, *exID, *setIdx, weight, reps, rir, completed, rest).Scan(&id)
	if err != nil {
		internalErr(w, err)
		return
	}
	jsonWrite(w, http.StatusCreated, map[string]any{"id": id})
}

func SetsPatch(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/sets/")
	if idStr == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	setID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || setID <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// ownership: pega session
	var sessionID int64
	if err := sessionsDB.QueryRow(`SELECT session_id FROM workout_sets WHERE id=$1`, setID).Scan(&sessionID); err != nil {
		http.NotFound(w, r)
		return
	}
	if uid := GetUserID(r); uid != "" {
		var ok bool
		if err := sessionsDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM workout_sessions WHERE id=$1 AND user_id=$2)`, sessionID, uid).Scan(&ok); err != nil || !ok {
			http.NotFound(w, r)
			return
		}
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		badRequest(w, "invalid json")
		return
	}

	type field struct {
		col string
		val any
	}
	var sets []field

	if v, ok := body["weight_kg"]; ok {
		sets = append(sets, field{"weight_kg", v})
	} else if v, ok := body["carga_kg"]; ok {
		sets = append(sets, field{"weight_kg", v})
	}
	if v, ok := body["reps"]; ok {
		sets = append(sets, field{"reps", v})
	} else if v, ok := body["repeticoes"]; ok {
		sets = append(sets, field{"reps", v})
	}
	if v, ok := body["set_index"]; ok {
		sets = append(sets, field{"set_index", v})
	} else if v, ok := body["series"]; ok {
		sets = append(sets, field{"set_index", v})
	}
	if v, ok := body["rir"]; ok {
		sets = append(sets, field{"rir", v})
	}
	if v, ok := body["completed"]; ok {
		sets = append(sets, field{"completed", v})
	}
	if v, ok := body["rest_sec"]; ok {
		sets = append(sets, field{"rest_sec", v})
	}

	if len(sets) == 0 {
		badRequest(w, "no updatable fields")
		return
	}

	args := []any{}
	sqlSet := make([]string, 0, len(sets))
	for i, f := range sets {
		sqlSet = append(sqlSet, f.col+" = $"+strconv.Itoa(i+1))
		args = append(args, f.val)
	}
	args = append(args, setID)
	q := `UPDATE workout_sets SET ` + strings.Join(sqlSet, ", ") + ` WHERE id = $` + strconv.Itoa(len(args)) + ` RETURNING id`
	var id int64
	if err := sessionsDB.QueryRow(q, args...).Scan(&id); err != nil {
		internalErr(w, err)
		return
	}
	jsonWrite(w, http.StatusOK, map[string]any{"id": id})
}

func SetsDelete(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/sets/")
	if idStr == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	setID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || setID <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// ownership
	var sessionID int64
	if err := sessionsDB.QueryRow(`SELECT session_id FROM workout_sets WHERE id=$1`, setID).Scan(&sessionID); err != nil {
		http.NotFound(w, r)
		return
	}
	if uid := GetUserID(r); uid != "" {
		var ok bool
		if err := sessionsDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM workout_sessions WHERE id=$1 AND user_id=$2)`, sessionID, uid).Scan(&ok); err != nil || !ok {
			http.NotFound(w, r)
			return
		}
	}
	if _, err := sessionsDB.Exec(`DELETE FROM workout_sets WHERE id=$1`, setID); err != nil {
		internalErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

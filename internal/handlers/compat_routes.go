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
// Guarda por dono: ($user=” OR ws.user_id IS NULL OR ws.user_id=$user)
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
	// guarda de dono (ou compat com NULL)
	whereParts = append(whereParts, "($"+itoa(argIdx)+" = '' OR ws.user_id IS NULL OR ws.user_id = $"+itoa(argIdx)+")")
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

	// from/to (RFC3339)
	parseTS := func(s string) (time.Time, bool) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, ok := parseTS(v); ok {
			whereParts = append(whereParts, "ws.session_at >= $"+itoa(argIdx))
			args = append(args, t)
			argIdx++
		} else {
			badRequest(w, "invalid from (RFC3339)")
			return
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, ok := parseTS(v); ok {
			whereParts = append(whereParts, "ws.session_at <= $"+itoa(argIdx))
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
SELECT ws.id, ws.treino_id, ws.session_at, COALESCE(ws.notes,''), 
       COALESCE(ws.completed,false), ws.duration_min, ws.rpe_session, ws.user_id,
       ws.created_at, ws.updated_at
FROM workout_sessions ws
` + whereSQL + `
ORDER BY ws.session_at DESC, ws.id DESC
LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := sessionsDB.Query(listQ, args...)
	if err != nil {
		internalErr(w, err)
		return
	}
	defer rows.Close()

	type rowOut struct {
		ID          int64     `json:"id"`
		TreinoID    int64     `json:"treino_id"`
		SessionAt   time.Time `json:"session_at"`
		Notes       string    `json:"notes"`
		Completed   bool      `json:"completed"`
		DurationMin *int64    `json:"duration_min,omitempty"`
		RPESession  *int64    `json:"rpe_session,omitempty"`
		UserID      *string   `json:"user_id,omitempty"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	items := make([]rowOut, 0, pageSize)
	for rows.Next() {
		var (
			r  rowOut
			dn sql.NullInt64
			rn sql.NullInt64
			un sql.NullString
		)
		if err := rows.Scan(
			&r.ID, &r.TreinoID, &r.SessionAt, &r.Notes, &r.Completed,
			&dn, &rn, &un, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			internalErr(w, err)
			return
		}
		if dn.Valid {
			r.DurationMin = &dn.Int64
		}
		if rn.Valid {
			r.RPESession = &rn.Int64
		}
		if un.Valid {
			u := un.String
			r.UserID = &u
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
SELECT ws.id, ws.treino_id, ws.session_at, COALESCE(ws.notes,''), 
       COALESCE(ws.completed,false), ws.duration_min, ws.rpe_session, ws.user_id,
       ws.created_at, ws.updated_at
FROM workout_sessions ws
WHERE ws.id = $1
  AND ($2 = '' OR ws.user_id IS NULL OR ws.user_id = $2)
`
	var (
		out struct {
			ID          int64     `json:"id"`
			TreinoID    int64     `json:"treino_id"`
			SessionAt   time.Time `json:"session_at"`
			Notes       string    `json:"notes"`
			Completed   bool      `json:"completed"`
			DurationMin *int64    `json:"duration_min,omitempty"`
			RPESession  *int64    `json:"rpe_session,omitempty"`
			UserID      *string   `json:"user_id,omitempty"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		}
		dn sql.NullInt64
		rn sql.NullInt64
		un sql.NullString
	)
	err = sessionsDB.QueryRow(q, id, userID).Scan(
		&out.ID, &out.TreinoID, &out.SessionAt, &out.Notes, &out.Completed,
		&dn, &rn, &un, &out.CreatedAt, &out.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalErr(w, err)
		return
	}
	if dn.Valid {
		out.DurationMin = &dn.Int64
	}
	if rn.Valid {
		out.RPESession = &rn.Int64
	}
	if un.Valid {
		u := un.String
		out.UserID = &u
	}
	jsonWrite(w, http.StatusOK, out)
}

// ============
// Overload GET
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
//  - GET/POST   /api/sessions/{id}/sets
//  - PATCH/DELETE /api/sets/{id}
// ==========================

func SetsList(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	setsCollectionCompat(sessionsDB).ServeHTTP(w, r)
}

func SetsCreate(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	setsCollectionCompat(sessionsDB).ServeHTTP(w, r)
}

func SetsPatch(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	setItemCompat(sessionsDB).ServeHTTP(w, r)
}

func SetsDelete(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	setItemCompat(sessionsDB).ServeHTTP(w, r)
}

// ===========================
// Implementações compatíveis
// ===========================

// coleção: GET lista de sets por sessão / POST cria set em sessão
// Esperado: /api/sessions/{id}/sets
func setsCollectionCompat(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parts: ["api","sessions","{id}","sets"]
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 4 {
			badRequest(w, "missing session id")
			return
		}
		sessionID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || sessionID <= 0 {
			badRequest(w, "invalid session id")
			return
		}

		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query(`
				SELECT id, session_id, exercicio_id, series, repeticoes,
				       COALESCE(carga_kg,0), COALESCE(rir,0),
				       completed, COALESCE(notes,'')
				FROM workout_sets
				WHERE session_id = $1
				ORDER BY id ASC
			`, sessionID)
			if err != nil {
				internalErr(w, err)
				return
			}
			defer rows.Close()

			items := make([]map[string]any, 0, 32)
			for rows.Next() {
				var (
					id, sessID, exID int64
					series, reps     int
					carga            float64
					rir              int
					completed        bool
					notes            string
				)
				if err := rows.Scan(&id, &sessID, &exID, &series, &reps, &carga, &rir, &completed, &notes); err != nil {
					internalErr(w, err)
					return
				}
				items = append(items, map[string]any{
					"id":           id,
					"session_id":   sessID,
					"exercicio_id": exID,
					"series":       series,
					"repeticoes":   reps,
					"carga_kg":     carga,
					"rir":          rir,
					"completed":    completed,
					"notes":        notes,
				})
			}
			jsonWrite(w, http.StatusOK, map[string]any{"items": items})
			return

		case http.MethodPost:
			// Dono do request (para guarda por owner na sessão)
			userID := strings.TrimSpace(GetUserID(r))

			var in struct {
				ExercicioID int64    `json:"exercicio_id"`
				Series      int      `json:"series"`
				Repeticoes  int      `json:"repeticoes"`
				CargaKg     *float64 `json:"carga_kg,omitempty"`
				RIR         *int     `json:"rir,omitempty"`
				Completed   *bool    `json:"completed,omitempty"`
				Notes       *string  `json:"notes,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				badRequest(w, "invalid json")
				return
			}
			if in.ExercicioID <= 0 || in.Series <= 0 || in.Repeticoes <= 0 {
				badRequest(w, "exercicio_id, series, repeticoes required")
				return
			}

			var (
				carga interface{}
				rir   interface{}
				comp  bool
				notes string
			)
			if in.CargaKg != nil {
				carga = *in.CargaKg
			} else {
				carga = nil
			}
			if in.RIR != nil {
				rir = *in.RIR
			} else {
				rir = nil
			}
			if in.Completed != nil {
				comp = *in.Completed
			}
			if in.Notes != nil {
				notes = *in.Notes
			}

			q := `
INSERT INTO workout_sets (session_id, exercicio_id, series, repeticoes, carga_kg, rir, completed, notes)
SELECT $1, $2, $3, $4, $5, $6, $7, $8
WHERE EXISTS (
  SELECT 1 FROM workout_sessions ws
  WHERE ws.id = $1
    AND ($9 = '' OR ws.user_id IS NULL OR ws.user_id = $9)
)
RETURNING id, session_id, exercicio_id, series, repeticoes, carga_kg, rir, completed, notes;
`
			var (
				id, sessID, exID int64
				series, reps     int
				cargaOut         sql.NullFloat64
				rirOut           sql.NullInt64
				completed        bool
				notesOut         sql.NullString
			)
			err := db.QueryRow(q,
				sessionID, in.ExercicioID, in.Series, in.Repeticoes,
				carga, rir, comp, notes,
				userID,
			).Scan(&id, &sessID, &exID, &series, &reps, &cargaOut, &rirOut, &completed, &notesOut)
			if err != nil {
				http.NotFound(w, r) // sessão não existe ou não pertence ao user
				return
			}

			resp := map[string]any{
				"id":           id,
				"session_id":   sessID,
				"exercicio_id": exID,
				"series":       series,
				"repeticoes":   reps,
				"completed":    completed,
			}
			if cargaOut.Valid {
				resp["carga_kg"] = cargaOut.Float64
			}
			if rirOut.Valid {
				resp["rir"] = int(rirOut.Int64)
			}
			if notesOut.Valid {
				resp["notes"] = notesOut.String
			}
			jsonWrite(w, http.StatusCreated, resp)
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
}

// item: PATCH/DELETE /api/sets/{id}
func setItemCompat(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			badRequest(w, "missing set id")
			return
		}
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || id <= 0 {
			badRequest(w, "invalid set id")
			return
		}
		userID := strings.TrimSpace(GetUserID(r))

		switch r.Method {
		case http.MethodPatch:
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
				"completed":  true,
				"rir":        true,
				"carga_kg":   true,
				"repeticoes": true,
				"notes":      true,
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
				case "completed":
					b, ok := v.(bool)
					if !ok {
						badRequest(w, "completed must be bool")
						return
					}
					setParts = append(setParts, "completed = $"+itoa(argIdx))
					args = append(args, b)
					argIdx++
				case "rir":
					iv, ok := toInt64(v)
					if !ok {
						badRequest(w, "rir must be int")
						return
					}
					setParts = append(setParts, "rir = $"+itoa(argIdx))
					args = append(args, iv)
					argIdx++
				case "carga_kg":
					fv, ok := toFloat64(v)
					if !ok {
						badRequest(w, "carga_kg must be number")
						return
					}
					setParts = append(setParts, "carga_kg = $"+itoa(argIdx))
					args = append(args, fv)
					argIdx++
				case "repeticoes":
					iv, ok := toInt64(v)
					if !ok {
						badRequest(w, "repeticoes must be int")
						return
					}
					setParts = append(setParts, "repeticoes = $"+itoa(argIdx))
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
				}
			}

			q := `
UPDATE workout_sets AS s
SET ` + strings.Join(setParts, ", ") + `
FROM workout_sessions AS ws
WHERE s.id = $` + itoa(argIdx) + `
  AND ws.id = s.session_id
  AND ($` + itoa(argIdx+1) + ` = '' OR ws.user_id IS NULL OR ws.user_id = $` + itoa(argIdx+1) + `)
`
			args = append(args, id, userID)

			res, err := db.Exec(q, args...)
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
			return

		case http.MethodDelete:
			q := `
DELETE FROM workout_sets s
USING workout_sessions ws
WHERE s.id = $1
  AND ws.id = s.session_id
  AND ($2 = '' OR ws.user_id IS NULL OR ws.user_id = $2)
`
			res, err := db.Exec(q, id, userID)
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
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
}

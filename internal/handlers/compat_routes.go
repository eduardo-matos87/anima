package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

var sessionsDB *sql.DB

// TreinosItem ainda não implementado de verdade — stub 501 para compatibilidade
func TreinosItem(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonWrite(w, http.StatusNotImplemented, map[string]any{
			"error":   "not_implemented",
			"message": "TreinosItem será implementado após consolidar os handlers de treinos.",
		})
	})
}

// ====== injeção de DB usada pelo main.go ======
func SetSessionsDB(db *sql.DB) { sessionsDB = db }

// =======================
// Wrappers no estilo (w,r)
// =======================

// /api/sessions/list
func SessionsList(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	ListSessions(sessionsDB).ServeHTTP(w, r)
}

// POST /api/sessions
func SessionsCreate(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	CreateSession(sessionsDB).ServeHTTP(w, r)
}

// GET /api/sessions/{id}
func SessionsGet(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	GetWorkoutSession(sessionsDB).ServeHTTP(w, r)
}

// PATCH /api/sessions/update/{id}
func SessionsPatch(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	UpdateDeleteSession(sessionsDB).ServeHTTP(w, r)
}

// DELETE /api/sessions/update/{id}
func SessionsDelete(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	UpdateDeleteSession(sessionsDB).ServeHTTP(w, r)
}

// POST /api/overload/suggest (wrapper legacy GET também aponta aqui no main)
func NextLoad(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	OverloadSuggest(sessionsDB).ServeHTTP(w, r)
}

// ==========================
// Sets compat com seu main:
//  - GET/POST /api/sessions/{id}/sets
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
// Esperado pelo main: /api/sessions/{id}/sets
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
				SELECT id, session_id, exercicio_id, series, repeticoes, COALESCE(carga_kg,0), COALESCE(rir,0), completed, COALESCE(notes,'')
				FROM workout_sets
				WHERE session_id = $1
				ORDER BY id ASC
			`, sessionID)
			if err != nil {
				internalErr(w, err)
				return
			}
			defer rows.Close()

			out := []SessionSet{}
			for rows.Next() {
				var s SessionSet
				if err := rows.Scan(&s.ID, &s.SessionID, &s.ExercicioID, &s.Series, &s.Repeticoes, &s.CargaKg, &s.RIR, &s.Completed, &s.Notes); err != nil {
					internalErr(w, err)
					return
				}
				out = append(out, s)
			}
			jsonWrite(w, http.StatusOK, map[string]any{"items": out})
			return

		case http.MethodPost:
			var in SessionSet
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				badRequest(w, "invalid json")
				return
			}
			in.SessionID = sessionID
			err := db.QueryRow(`
				INSERT INTO workout_sets (session_id, exercicio_id, series, repeticoes, carga_kg, rir, completed, notes)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
				RETURNING id
			`, in.SessionID, in.ExercicioID, in.Series, in.Repeticoes, in.CargaKg, in.RIR, in.Completed, in.Notes).Scan(&in.ID)
			if err != nil {
				internalErr(w, err)
				return
			}
			jsonWrite(w, http.StatusCreated, in)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// item: PATCH/DELETE /api/sets/{id}
func setItemCompat(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parts: ["api","sets","{id}"]
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			badRequest(w, "missing set id")
			return
		}
		setID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || setID <= 0 {
			badRequest(w, "invalid set id")
			return
		}

		switch r.Method {
		case http.MethodPatch:
			var in SessionSet
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				badRequest(w, "invalid json")
				return
			}
			// Atualiza apenas campos enviados
			q := `UPDATE workout_sets SET `
			args := []any{}
			i := 1
			if in.ExercicioID != 0 {
				q += `exercicio_id = $` + fmtInt(i) + `, `
				args = append(args, in.ExercicioID)
				i++
			}
			if in.Series != 0 {
				q += `series = $` + fmtInt(i) + `, `
				args = append(args, in.Series)
				i++
			}
			if in.Repeticoes != 0 {
				q += `repeticoes = $` + fmtInt(i) + `, `
				args = append(args, in.Repeticoes)
				i++
			}
			q += `carga_kg = COALESCE($` + fmtInt(i) + `, carga_kg), `
			args = append(args, nullFloat(in.CargaKg))
			i++
			q += `rir = COALESCE($` + fmtInt(i) + `, rir), `
			args = append(args, nullInt(in.RIR))
			i++
			q += `completed = COALESCE($` + fmtInt(i) + `, completed), `
			args = append(args, nullBoolPtr(in.Completed))
			i++
			q += `notes = COALESCE($` + fmtInt(i) + `, notes) `
			args = append(args, nullString(in.Notes))
			i++
			q += `WHERE id = $` + fmtInt(i)

			args = append(args, setID)
			if _, err := db.Exec(q, args...); err != nil {
				internalErr(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return

		case http.MethodDelete:
			if _, err := db.Exec(`DELETE FROM workout_sets WHERE id=$1`, setID); err != nil {
				internalErr(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

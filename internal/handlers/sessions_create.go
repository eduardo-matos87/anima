package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type createSessionReq struct {
	TreinoID  int64        `json:"treino_id"`
	SessionAt time.Time    `json:"session_at"`
	Notes     string       `json:"notes,omitempty"`
	Sets      []SessionSet `json:"sets,omitempty"`
}

// POST /api/sessions
func CreateSession(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in createSessionReq
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			badRequest(w, "invalid json")
			return
		}
		if err := validateSessionPayload(in.TreinoID, in.SessionAt); err != nil {
			badRequest(w, err.Error())
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			internalErr(w, err)
			return
		}
		var sid int64
		err = tx.QueryRow(`
			INSERT INTO workout_sessions (treino_id, session_at, notes)
			VALUES ($1,$2,$3)
			RETURNING id
		`, in.TreinoID, in.SessionAt, in.Notes).Scan(&sid)
		if err != nil {
			_ = tx.Rollback()
			internalErr(w, err)
			return
		}
		// sets opcionais
		if len(in.Sets) > 0 {
			stmt, err := tx.Prepare(`
			  INSERT INTO workout_sets (session_id, exercicio_id, series, repeticoes, carga_kg, rir, completed, notes)
			  VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`)
			if err != nil {
				_ = tx.Rollback()
				internalErr(w, err)
				return
			}
			defer stmt.Close()
			for _, s := range in.Sets {
				if _, err := stmt.Exec(sid, s.ExercicioID, s.Series, s.Repeticoes, s.CargaKg, s.RIR, s.Completed, s.Notes); err != nil {
					_ = tx.Rollback()
					internalErr(w, err)
					return
				}
			}
		}
		if err := tx.Commit(); err != nil {
			internalErr(w, err)
			return
		}
		// ✅ aqui era writeJSON; trocamos por jsonWrite (helper deste pacote)
		jsonWrite(w, http.StatusCreated, map[string]any{"id": sid})
	})
}

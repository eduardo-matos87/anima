package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
)

type overloadReq struct {
	ExercicioID int64 `json:"exercicio_id"`
	Window      int   `json:"window,omitempty"` // 3..12 (default 5)
}
type overloadResp struct {
	SuggestedCargaKg float64 `json:"suggested_carga_kg"`
	SuggestedReps    int     `json:"suggested_repeticoes"`
	Rationale        string  `json:"rationale"`
	AvgCargaKg       float64 `json:"avg_carga_kg"`
	AvgRIR           float64 `json:"avg_rir"`
	SampleCount      int     `json:"sample_count"`
}

// POST /api/overload/suggest
// GET  /api/overload/suggest?exercicio_id=10&window=5
// GET  /api/suggestions/next-load?exercicio_id=10&window=5 (legacy)
func OverloadSuggest(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in overloadReq

		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			idStr := q.Get("exercicio_id")
			if idStr == "" {
				badRequest(w, "missing exercicio_id")
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil || id <= 0 {
				badRequest(w, "invalid exercicio_id")
				return
			}
			in.ExercicioID = id
			if win := q.Get("window"); win != "" {
				if wv, err := strconv.Atoi(win); err == nil {
					in.Window = wv
				}
			}
		default: // POST
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ExercicioID <= 0 {
				badRequest(w, "invalid json or exercicio_id")
				return
			}
		}

		if in.Window < 3 || in.Window > 12 {
			in.Window = 5
		}

		var (
			avgCarga float64
			avgRIR   float64
			n        int
		)

		// window==12 → view otimizada; senão subquery
		if in.Window == 12 {
			row := db.QueryRow(`
				SELECT
					COALESCE(avg_carga_kg::float8, 0),
					COALESCE(avg_rir::float8, 1.5),
					COALESCE(sample_count, 0)
				FROM workout_overload_stats12
				WHERE exercicio_id = $1
			`, in.ExercicioID)
			if err := row.Scan(&avgCarga, &avgRIR, &n); err != nil {
				internalErr(w, err)
				return
			}
		} else {
			row := db.QueryRow(`
			   SELECT
			     COALESCE(AVG(carga_kg), 0) AS avg_carga,
			     COALESCE(AVG(rir), 1.5)    AS avg_rir,
			     COUNT(*)                    AS n
			   FROM (
			     SELECT carga_kg, rir
			     FROM workout_sets
			     WHERE exercicio_id = $1
			       AND completed = TRUE
			     ORDER BY id DESC
			     LIMIT $2
			   ) s
			`, in.ExercicioID, in.Window)
			if err := row.Scan(&avgCarga, &avgRIR, &n); err != nil {
				internalErr(w, err)
				return
			}
		}

		// sem amostras: resposta neutra
		if n == 0 {
			resp := overloadResp{
				SuggestedCargaKg: 0,
				SuggestedReps:    10,
				Rationale:        "sem histórico concluído para este exercício",
				AvgCargaKg:       0,
				AvgRIR:           1.5,
				SampleCount:      0,
			}
			jsonWrite(w, http.StatusOK, resp)
			insertOverloadLog(db, r, in, resp)
			return
		}

		// regra adaptativa
		sugCarga := roundTo(avgCarga, 0.5)
		sugReps := 10
		rationale := "RIR moderado, manter carga"
		switch {
		case avgRIR >= 2.5:
			sugCarga = roundTo(avgCarga+5.0, 0.5)
			rationale = "RIR muito alto, sugere +5kg"
		case avgRIR >= 1.8:
			sugCarga = roundTo(avgCarga+2.5, 0.5)
			rationale = "RIR alto, sugere +2.5kg"
		case avgRIR <= 0.5:
			sugReps = 8
			rationale = "RIR baixo, manter carga e reduzir reps"
		}

		resp := overloadResp{
			SuggestedCargaKg: sugCarga,
			SuggestedReps:    sugReps,
			Rationale:        rationale,
			AvgCargaKg:       roundTo(avgCarga, 0.5),
			AvgRIR:           avgRIR,
			SampleCount:      n,
		}
		jsonWrite(w, http.StatusOK, resp)
		insertOverloadLog(db, r, in, resp)
	})
}

func roundTo(v, step float64) float64 {
	return math.Round(v/step) * step
}

// ===== logging auxiliar (defensivo) =====

func insertOverloadLog(db *sql.DB, r *http.Request, in overloadReq, out overloadResp) {
	userID := r.Header.Get("X-User-ID")
	ip := clientIP(r)
	ua := r.UserAgent()

	// garante a tabela (caso a migração ainda não tenha sido aplicada)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS overload_suggestions_log (
		  id                   BIGSERIAL PRIMARY KEY,
		  requested_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		  user_id              TEXT,
		  ip                   INET,
		  user_agent           TEXT,
		  exercicio_id         BIGINT NOT NULL,
		  window_size          INT NOT NULL,
		  avg_carga_kg         NUMERIC(10,2),
		  avg_rir              NUMERIC(10,2),
		  sample_count         INT,
		  suggested_carga_kg   NUMERIC(10,2),
		  suggested_repeticoes INT,
		  rationale            TEXT
		);
	`)
	if err != nil {
		log.Printf("[overload_log] create table failed: %v", err)
		return
	}
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_overload_log_exercicio_at
		  ON overload_suggestions_log (exercicio_id, requested_at DESC);
	`)
	if err != nil {
		log.Printf("[overload_log] create index exercicio_at failed: %v", err)
	}
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_overload_log_user_at
		  ON overload_suggestions_log (user_id, requested_at DESC);
	`)
	if err != nil {
		log.Printf("[overload_log] create index user_at failed: %v", err)
	}

	// insert best-effort
	_, err = db.Exec(`
		INSERT INTO overload_suggestions_log
		  (requested_at, user_id, ip, user_agent, exercicio_id, window_size,
		   avg_carga_kg, avg_rir, sample_count, suggested_carga_kg, suggested_repeticoes, rationale)
		VALUES (NOW(), $1, $2, $3, $4, $5,
		        $6, $7, $8, $9, $10, $11)
	`, nullString(userID), nullString(ip), nullString(ua),
		in.ExercicioID, in.Window,
		out.AvgCargaKg, out.AvgRIR, out.SampleCount, out.SuggestedCargaKg, out.SuggestedReps, out.Rationale)
	if err != nil {
		log.Printf("[overload_log] insert failed: %v", err)
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

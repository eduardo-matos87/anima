package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"time"
)

// AdminOverloadRefresh executa o REFRESH da MV de overload com melhores práticas:
// - Apenas POST
// - Auth via header X-Admin-Token (se ADMIN_TOKEN estiver setado)
// - Tenta CONCURRENTLY; se falhar, tenta normal
// - Se a MV não existir, cria (WITH NO DATA) + índice e faz o primeiro refresh
// - Atualiza admin_metadata.overload_last_refresh
func AdminOverloadRefresh(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enforce POST
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Auth simples
		want := os.Getenv("ADMIN_TOKEN")
		got := r.Header.Get("X-Admin-Token")
		if want != "" && got != want {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		mode := "concurrent"
		createdMV := false

		// 1) tenta CONCURRENTLY
		if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY workout_overload_stats12_mv`); err != nil {
			// 2) fallback: tenta normal
			mode = "full"
			if _, err2 := db.Exec(`REFRESH MATERIALIZED VIEW workout_overload_stats12_mv`); err2 != nil {
				// 3) se falhou, tenta criar MV e índice e depois refrescar (primeiro populate)
				// (pressupõe que a view workout_sets_recent12 já exista; criada pela migration 023)
				if _, err3 := db.Exec(`
					CREATE MATERIALIZED VIEW IF NOT EXISTS workout_overload_stats12_mv AS
					SELECT
					  exercicio_id,
					  COALESCE(AVG(carga_kg), 0)::numeric(10,2) AS avg_carga_kg,
					  COALESCE(AVG(rir), 1.5)::numeric(10,2)     AS avg_rir,
					  COUNT(*)                                    AS sample_count
					FROM workout_sets_recent12
					GROUP BY exercicio_id
					WITH NO DATA;
				`); err3 == nil {
					createdMV = true
					_, _ = db.Exec(`
						CREATE UNIQUE INDEX IF NOT EXISTS workout_overload_stats12_mv_pk
						ON workout_overload_stats12_mv (exercicio_id);
					`)
					// primeiro populate (sem concurrently)
					if _, err4 := db.Exec(`REFRESH MATERIALIZED VIEW workout_overload_stats12_mv`); err4 != nil {
						jsonWrite(w, http.StatusInternalServerError, map[string]any{
							"error":   "refresh_failed",
							"message": err4.Error(),
							"step":    "create_mv_and_refresh",
						})
						return
					}
				} else {
					jsonWrite(w, http.StatusInternalServerError, map[string]any{
						"error":   "refresh_failed",
						"message": err2.Error(),
						"step":    "fallback_full",
					})
					return
				}
			}
		}

		// 4) marca metadata (idempotente)
		_, _ = db.Exec(`
			CREATE TABLE IF NOT EXISTS admin_metadata (
			  key TEXT PRIMARY KEY,
			  value TEXT,
			  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);
		`)
		_, _ = db.Exec(`
			INSERT INTO admin_metadata(key, value, updated_at)
			VALUES ('overload_last_refresh', NOW()::text, NOW())
			ON CONFLICT (key) DO UPDATE
			  SET value = EXCLUDED.value, updated_at = NOW();
		`)

		jsonWrite(w, http.StatusOK, map[string]any{
			"ok":           true,
			"mode":         mode,      // "concurrent" ou "full"
			"created_mv":   createdMV, // true se a MV foi criada aqui
			"refreshed_at": time.Now().UTC().Format(time.RFC3339Nano),
		})
	})
}

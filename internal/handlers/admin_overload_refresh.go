package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type refreshResp struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func OverloadRefreshMV(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(refreshResp{Status: "method_not_allowed"})
			return
		}

		// Garante MV por usuário e índice único
		if _, err := db.Exec(`
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_matviews
    WHERE schemaname = 'public' AND matviewname = 'workout_overload_stats12_user_mv'
  ) THEN
    CREATE MATERIALIZED VIEW workout_overload_stats12_user_mv AS
    SELECT
      user_id,
      exercicio_id,
      COALESCE(AVG(weight_kg), 0)::numeric(10,2) AS avg_carga_kg,
      COALESCE(AVG(rir), 1.5)::numeric(10,2)     AS avg_rir,
      COUNT(*)                                   AS sample_count
    FROM workout_sets_recent12_user
    GROUP BY user_id, exercicio_id
    WITH NO DATA;
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_overload12_user_mv_pk
  ON workout_overload_stats12_user_mv (user_id, exercicio_id);
`); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(refreshResp{Status: "ensure_failed", Error: err.Error()})
			return
		}

		// Tenta CONCURRENTLY; se falhar, tenta normal.
		if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY workout_overload_stats12_user_mv`); err != nil {
			if _, err2 := db.Exec(`REFRESH MATERIALIZED VIEW workout_overload_stats12_user_mv`); err2 != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				_ = json.NewEncoder(w).Encode(refreshResp{Status: "refresh_failed", Error: err2.Error()})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(refreshResp{Status: "refreshed"})
	})
}

// Back-compat com main.go antigo
func AdminOverloadRefresh(db *sql.DB) http.Handler {
	return OverloadRefreshMV(db)
}

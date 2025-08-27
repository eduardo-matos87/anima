package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"time"
)

// AdminOverloadStats retorna agregados dos logs de overload.
// Auth: X-Admin-Token (se ADMIN_TOKEN setado).
// GET /api/admin/overload/stats?group=exercicio|day|hour&exercicio_id=10&from=RFC3339&to=RFC3339&limit=30
func AdminOverloadStats(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// auth
		want := os.Getenv("ADMIN_TOKEN")
		got := r.Header.Get("X-Admin-Token")
		if want != "" && got != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
		group := q.Get("group")
		if group == "" {
			group = "exercicio"
		}
		// filtros
		var where = "WHERE 1=1"
		var args []any
		i := 1

		if s := q.Get("exercicio_id"); s != "" {
			id, err := strconv.ParseInt(s, 10, 64)
			if err != nil || id <= 0 {
				badRequest(w, "invalid exercicio_id")
				return
			}
			where += " AND exercicio_id = $" + fmtInt(i)
			args = append(args, id)
			i++
		}
		if s := q.Get("from"); s != "" {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				badRequest(w, "invalid from (RFC3339)")
				return
			}
			where += " AND requested_at >= $" + fmtInt(i)
			args = append(args, t)
			i++
		}
		if s := q.Get("to"); s != "" {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				badRequest(w, "invalid to (RFC3339)")
				return
			}
			where += " AND requested_at <= $" + fmtInt(i)
			args = append(args, t)
			i++
		}

		limit := clampInt(atoiDefault(q.Get("limit"), 30), 1, 1000)

		switch group {
		case "exercicio":
			sqlStr := `
				SELECT
				  exercicio_id,
				  COUNT(*) AS requests,
				  COALESCE(AVG(suggested_carga_kg::float8),0),
				  COALESCE(AVG(avg_carga_kg::float8),0),
				  COALESCE(AVG(avg_rir::float8),0),
				  COALESCE(SUM(sample_count),0),
				  MAX(requested_at) AS last_requested_at
				FROM overload_suggestions_log
				` + where + `
				GROUP BY exercicio_id
				ORDER BY requests DESC
				LIMIT $` + fmtInt(i)
			args = append(args, limit)

			type row struct {
				ExercicioID       int64     `json:"exercicio_id"`
				Requests          int64     `json:"requests"`
				AvgSuggestedCarga float64   `json:"avg_suggested_carga_kg"`
				AvgOfAvgCarga     float64   `json:"avg_of_avg_carga_kg"`
				AvgRIR            float64   `json:"avg_rir"`
				TotalSamples      int64     `json:"total_samples"`
				LastRequestedAt   time.Time `json:"last_requested_at"`
			}

			rows, err := db.Query(sqlStr, args...)
			if err != nil {
				internalErr(w, err)
				return
			}
			defer rows.Close()

			var out []row
			for rows.Next() {
				var it row
				if err := rows.Scan(
					&it.ExercicioID, &it.Requests, &it.AvgSuggestedCarga, &it.AvgOfAvgCarga,
					&it.AvgRIR, &it.TotalSamples, &it.LastRequestedAt,
				); err != nil {
					internalErr(w, err)
					return
				}
				out = append(out, it)
			}
			jsonWrite(w, http.StatusOK, map[string]any{
				"group": "exercicio",
				"items": out,
				"limit": limit,
			})
			return

		case "day", "hour":
			trunc := "day"
			if group == "hour" {
				trunc = "hour"
			}
			sqlStr := `
				SELECT
				  date_trunc('` + trunc + `', requested_at) AS bucket,
				  COUNT(*) AS requests,
				  COALESCE(AVG(suggested_carga_kg::float8),0),
				  COALESCE(AVG(avg_rir::float8),0)
				FROM overload_suggestions_log
				` + where + `
				GROUP BY bucket
				ORDER BY bucket DESC
				LIMIT $` + fmtInt(i)
			args = append(args, limit)

			type row struct {
				Bucket              time.Time `json:"bucket"`
				Requests            int64     `json:"requests"`
				AvgSuggestedCargaKg float64   `json:"avg_suggested_carga_kg"`
				AvgRIR              float64   `json:"avg_rir"`
			}

			rows, err := db.Query(sqlStr, args...)
			if err != nil {
				internalErr(w, err)
				return
			}
			defer rows.Close()

			var out []row
			for rows.Next() {
				var it row
				if err := rows.Scan(&it.Bucket, &it.Requests, &it.AvgSuggestedCargaKg, &it.AvgRIR); err != nil {
					internalErr(w, err)
					return
				}
				out = append(out, it)
			}
			jsonWrite(w, http.StatusOK, map[string]any{
				"group": group,
				"items": out,
				"limit": limit,
			})
			return

		default:
			badRequest(w, "invalid group (use exercicio|day|hour)")
			return
		}
	})
}

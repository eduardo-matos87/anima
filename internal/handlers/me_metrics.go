package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MeMetrics: estatísticas pessoais a partir do overload_suggestions_log.
// GET /api/me/metrics?from=RFC3339&to=RFC3339&top=10
func MeMetrics(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(GetUserID(r))
		if userID == "" {
			http.Error(w, "unauthorized (missing user id)", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
		var (
			from time.Time
			to   time.Time
			err  error
		)
		if s := q.Get("from"); s != "" {
			from, err = time.Parse(time.RFC3339, s)
			if err != nil {
				badRequest(w, "invalid from (RFC3339)")
				return
			}
		} else {
			from = time.Now().AddDate(0, 0, -30) // 30 dias padrão
		}
		if s := q.Get("to"); s != "" {
			to, err = time.Parse(time.RFC3339, s)
			if err != nil {
				badRequest(w, "invalid to (RFC3339)")
				return
			}
		} else {
			to = time.Now()
		}

		topN := 10
		if t := q.Get("top"); t != "" {
			if n, err := strconv.Atoi(t); err == nil && n > 0 && n <= 100 {
				topN = n
			}
		}

		// Totais no período
		var totalReq, uniqEx int64
		var avgSug, avgRIR sql.NullFloat64
		var avgSamples sql.NullFloat64
		if err := db.QueryRow(`
			SELECT
			  COUNT(*) AS total_requests,
			  COUNT(DISTINCT exercicio_id) AS unique_exercises,
			  COALESCE(AVG(suggested_carga_kg::float8),0) AS avg_suggested_carga,
			  COALESCE(AVG(avg_rir::float8),0) AS avg_rir,
			  COALESCE(AVG(sample_count::float8),0) AS avg_samples
			FROM overload_suggestions_log
			WHERE user_id = $1 AND requested_at >= $2 AND requested_at <= $3
		`, userID, from, to).Scan(&totalReq, &uniqEx, &avgSug, &avgRIR, &avgSamples); err != nil {
			internalErr(w, err)
			return
		}

		// Janelas 7d/30d
		var last7, last30 int64
		if err := db.QueryRow(`
			SELECT COUNT(*) FROM overload_suggestions_log
			WHERE user_id = $1 AND requested_at >= NOW() - INTERVAL '7 days'
		`, userID).Scan(&last7); err != nil {
			internalErr(w, err)
			return
		}
		if err := db.QueryRow(`
			SELECT COUNT(*) FROM overload_suggestions_log
			WHERE user_id = $1 AND requested_at >= NOW() - INTERVAL '30 days'
		`, userID).Scan(&last30); err != nil {
			internalErr(w, err)
			return
		}

		// Top exercícios
		rows, err := db.Query(`
			SELECT
			  exercicio_id,
			  COUNT(*) AS requests,
			  COALESCE(AVG(suggested_carga_kg::float8),0) AS avg_suggested_carga_kg,
			  MAX(requested_at) AS last_requested_at
			FROM overload_suggestions_log
			WHERE user_id = $1 AND requested_at >= $2 AND requested_at <= $3
			GROUP BY exercicio_id
			ORDER BY requests DESC, last_requested_at DESC
			LIMIT $4
		`, userID, from, to, topN)
		if err != nil {
			internalErr(w, err)
			return
		}
		defer rows.Close()

		type topItem struct {
			ExercicioID       int64     `json:"exercicio_id"`
			Requests          int64     `json:"requests"`
			AvgSuggestedCarga float64   `json:"avg_suggested_carga_kg"`
			LastRequestedAt   time.Time `json:"last_requested_at"`
		}
		var tops []topItem
		for rows.Next() {
			var it topItem
			if err := rows.Scan(&it.ExercicioID, &it.Requests, &it.AvgSuggestedCarga, &it.LastRequestedAt); err != nil {
				internalErr(w, err)
				return
			}
			tops = append(tops, it)
		}

		out := map[string]any{
			"user_id": userID,
			"range": map[string]string{
				"from": from.UTC().Format(time.RFC3339),
				"to":   to.UTC().Format(time.RFC3339),
			},
			"totals": map[string]any{
				"requests":               totalReq,
				"unique_exercises":       uniqEx,
				"avg_suggested_carga_kg": nullF(avgSug),
				"avg_rir":                nullF(avgRIR),
				"avg_sample_count":       nullF(avgSamples),
			},
			"windows": map[string]any{
				"last_7d":  last7,
				"last_30d": last30,
			},
			"top_exercises": tops,
		}
		jsonWrite(w, http.StatusOK, out)
	})
}

func nullF(v sql.NullFloat64) float64 {
	if v.Valid {
		return v.Float64
	}
	return 0
}

package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// AdminOverloadLogs lista logs de overload com filtros e paginação.
// Segurança: exige header X-Admin-Token == ADMIN_TOKEN (se setado no ambiente).
// GET /api/admin/overload/logs?exercicio_id=10&user_id=abc&from=2025-08-01T00:00:00Z&to=2025-08-31T23:59:59Z&page=1&page_size=50
func AdminOverloadLogs(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth simples
		want := os.Getenv("ADMIN_TOKEN")
		got := r.Header.Get("X-Admin-Token")
		if want != "" && got != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()

		// filtros
		var (
			where = "WHERE 1=1"
			args  []any
			i     = 1
		)

		if s := q.Get("exercicio_id"); s != "" {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil && id > 0 {
				where += " AND exercicio_id = $" + fmtInt(i)
				args = append(args, id)
				i++
			} else {
				badRequest(w, "invalid exercicio_id")
				return
			}
		}

		if s := strings.TrimSpace(q.Get("user_id")); s != "" {
			where += " AND user_id = $" + fmtInt(i)
			args = append(args, s)
			i++
		}

		if s := q.Get("from"); s != "" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				where += " AND requested_at >= $" + fmtInt(i)
				args = append(args, t)
				i++
			} else {
				badRequest(w, "invalid from (use RFC3339)")
				return
			}
		}
		if s := q.Get("to"); s != "" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				where += " AND requested_at <= $" + fmtInt(i)
				args = append(args, t)
				i++
			} else {
				badRequest(w, "invalid to (use RFC3339)")
				return
			}
		}

		// paginação segura
		page := clampInt(parseIntOrDefault(q.Get("page"), 1), 1, 1_000_000_000)
		pageSize := clampInt(parseIntOrDefault(q.Get("page_size"), 50), 1, 200)
		offset := (page - 1) * pageSize

		// total
		var total int64
		if err := db.QueryRow("SELECT COUNT(*) FROM overload_suggestions_log "+where, args...).Scan(&total); err != nil {
			internalErr(w, err)
			return
		}

		// lista
		argsList := append(append([]any{}, args...), pageSize, offset)
		rows, err := db.Query(`
			SELECT
			  id, requested_at, user_id, ip, user_agent,
			  exercicio_id, window_size, avg_carga_kg, avg_rir, sample_count,
			  suggested_carga_kg, suggested_repeticoes, rationale
			FROM overload_suggestions_log
			`+where+`
			ORDER BY requested_at DESC, id DESC
			LIMIT $`+fmtInt(i)+` OFFSET $`+fmtInt(i+1), argsList...)
		if err != nil {
			internalErr(w, err)
			return
		}
		defer rows.Close()

		type item struct {
			ID                  int64     `json:"id"`
			RequestedAt         time.Time `json:"requested_at"`
			UserID              *string   `json:"user_id,omitempty"`
			IP                  *string   `json:"ip,omitempty"`
			UserAgent           *string   `json:"user_agent,omitempty"`
			ExercicioID         int64     `json:"exercicio_id"`
			WindowSize          int       `json:"window_size"`
			AvgCargaKg          *float64  `json:"avg_carga_kg,omitempty"`
			AvgRIR              *float64  `json:"avg_rir,omitempty"`
			SampleCount         *int      `json:"sample_count,omitempty"`
			SuggestedCargaKg    *float64  `json:"suggested_carga_kg,omitempty"`
			SuggestedRepeticoes *int      `json:"suggested_repeticoes,omitempty"`
			Rationale           *string   `json:"rationale,omitempty"`
		}

		out := []item{}
		for rows.Next() {
			var it item
			if err := rows.Scan(
				&it.ID, &it.RequestedAt, &it.UserID, &it.IP, &it.UserAgent,
				&it.ExercicioID, &it.WindowSize, &it.AvgCargaKg, &it.AvgRIR, &it.SampleCount,
				&it.SuggestedCargaKg, &it.SuggestedRepeticoes, &it.Rationale,
			); err != nil {
				internalErr(w, err)
				return
			}
			out = append(out, it)
		}

		jsonWrite(w, http.StatusOK, map[string]any{
			"items":      out,
			"page":       page,
			"page_size":  pageSize,
			"total_hint": total,
		})
	})
}

func parseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

package handlers

import (
	"database/sql"
	"encoding/csv"
	"net/http"
	"os"
	"strconv"
	"time"
)

// AdminOverloadExportCSV exporta logs como CSV (streaming).
// GET /api/admin/overload/export.csv?exercicio_id=10&user_id=abc&from=RFC3339&to=RFC3339&limit=10000
func AdminOverloadExportCSV(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// auth
		want := os.Getenv("ADMIN_TOKEN")
		got := r.Header.Get("X-Admin-Token")
		if want != "" && got != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
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
		if s := q.Get("user_id"); s != "" {
			where += " AND user_id = $" + fmtInt(i)
			args = append(args, s)
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

		limit := clampInt(atoiDefault(q.Get("limit"), 10000), 1, 100000)

		filename := "overload_logs_" + time.Now().UTC().Format("20060102T150405Z") + ".csv"
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

		writer := csv.NewWriter(w)
		defer writer.Flush()

		// header
		_ = writer.Write([]string{
			"id", "requested_at", "user_id", "ip", "user_agent",
			"exercicio_id", "window_size", "avg_carga_kg", "avg_rir", "sample_count",
			"suggested_carga_kg", "suggested_repeticoes", "rationale",
		})

		sqlStr := `
			SELECT
			  id, requested_at, COALESCE(user_id,''), COALESCE(CAST(ip AS TEXT),''), COALESCE(user_agent,''),
			  exercicio_id, window_size,
			  COALESCE(avg_carga_kg,0), COALESCE(avg_rir,0), COALESCE(sample_count,0),
			  COALESCE(suggested_carga_kg,0), COALESCE(suggested_repeticoes,0),
			  COALESCE(rationale,'')
			FROM overload_suggestions_log
			` + where + `
			ORDER BY id ASC
			LIMIT $` + fmtInt(i)
		args = append(args, limit)

		rows, err := db.Query(sqlStr, args...)
		if err != nil {
			internalErr(w, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id             int64
				reqAt          time.Time
				userID, ip, ua string
				exercicioID    int64
				windowSize     int
				avgCargaKg     float64
				avgRIR         float64
				sampleCount    int
				sugCarga       float64
				sugReps        int
				rationale      string
			)
			if err := rows.Scan(&id, &reqAt, &userID, &ip, &ua,
				&exercicioID, &windowSize, &avgCargaKg, &avgRIR, &sampleCount,
				&sugCarga, &sugReps, &rationale); err != nil {
				internalErr(w, err)
				return
			}
			rec := []string{
				strconv.FormatInt(id, 10),
				reqAt.UTC().Format(time.RFC3339),
				userID, ip, ua,
				strconv.FormatInt(exercicioID, 10),
				strconv.Itoa(windowSize),
				strconv.FormatFloat(avgCargaKg, 'f', -1, 64),
				strconv.FormatFloat(avgRIR, 'f', -1, 64),
				strconv.Itoa(sampleCount),
				strconv.FormatFloat(sugCarga, 'f', -1, 64),
				strconv.Itoa(sugReps),
				rationale,
			}
			_ = writer.Write(rec)
		}
	})
}

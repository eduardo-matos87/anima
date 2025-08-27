package handlers

import (
	"database/sql"
	"net/http"
)

// GET /api/sessions/list?treino_id=&from=&to=&page=&page_size=
func ListSessions(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := parseIntQuery(r, "page", 1)
		ps := parseIntQuery(r, "page_size", 20)
		if ps > 100 {
			ps = 100
		}
		offset := (page - 1) * ps

		treinoID := parseInt64Query(r, "treino_id", 0)
		from, hasFrom := parseTimeQuery(r, "from")
		to, hasTo := parseTimeQuery(r, "to")

		q := `
		  SELECT id, treino_id, session_at, COALESCE(notes,''), created_at, updated_at
		  FROM workout_sessions
		  WHERE 1=1`
		args := []any{}
		i := 1
		if treinoID > 0 {
			q += ` AND treino_id = $` + itoa(i)
			args = append(args, treinoID)
			i++
		}
		if hasFrom {
			q += ` AND session_at >= $` + itoa(i)
			args = append(args, from)
			i++
		}
		if hasTo {
			q += ` AND session_at <= $` + itoa(i)
			args = append(args, to)
			i++
		}
		q += ` ORDER BY session_at DESC LIMIT $` + itoa(i) + ` OFFSET $` + itoa(i+1)
		args = append(args, ps, offset)

		rows, err := db.Query(q, args...)
		if err != nil {
			internalErr(w, err)
			return
		}
		defer rows.Close()

		out := []Session{}
		for rows.Next() {
			var s Session
			if err := rows.Scan(&s.ID, &s.TreinoID, &s.SessionAt, &s.Notes, &s.CreatedAt, &s.UpdatedAt); err != nil {
				internalErr(w, err)
				return
			}
			out = append(out, s)
		}
		jsonWrite(w, http.StatusOK, map[string]any{
			"page":       page,
			"page_size":  ps,
			"items":      out,
			"total_hint": len(out),
		})
	})
}

func itoa(i int) string { return fmtInt(i) }

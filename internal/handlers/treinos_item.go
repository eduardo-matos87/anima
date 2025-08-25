package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// TreinosItem: GET /api/treinos/{id}
func TreinosItem(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		// /api /treinos /{id}
		if len(parts) < 3 {
			badRequest(w, "missing id")
			return
		}
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || id <= 0 {
			badRequest(w, "invalid id")
			return
		}

		rows, err := db.Query(`SELECT * FROM treinos WHERE id = $1`, id)
		if err != nil {
			internalErr(w, err)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			notFound(w)
			return
		}

		cols, err := rows.Columns()
		if err != nil {
			internalErr(w, err)
			return
		}
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			internalErr(w, err)
			return
		}

		obj := map[string]any{}
		for i, c := range cols {
			v := raw[i]
			// converte []byte -> string quando for texto
			if b, ok := v.([]byte); ok {
				obj[c] = string(b)
			} else {
				obj[c] = v
			}
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(obj); err != nil {
			internalErr(w, err)
			return
		}
	})
}

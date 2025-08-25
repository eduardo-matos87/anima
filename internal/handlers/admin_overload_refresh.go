package handlers

import (
	"database/sql"
	"net/http"
	"os"
)

func AdminOverloadRefresh(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// auth simples por header
		want := os.Getenv("ADMIN_TOKEN")
		got := r.Header.Get("X-Admin-Token")
		if want != "" && got != want {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY workout_overload_stats12_mv`); err != nil {
			jsonWrite(w, http.StatusNotFound, map[string]any{
				"error":   "refresh_failed",
				"message": err.Error(),
			})
			return
		}
		jsonWrite(w, http.StatusOK, map[string]any{"ok": true})
	})
}

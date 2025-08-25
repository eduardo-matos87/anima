package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

// GET /api/sessions/{id}
func GetWorkoutSession(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			badRequest(w, "missing id")
			return
		}
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || id <= 0 {
			badRequest(w, "invalid id")
			return
		}
		s, err := fetchSession(r.Context(), db, id)
		if err != nil {
			internalErr(w, err)
			return
		}
		if s == nil {
			notFound(w)
			return
		}
		jsonWrite(w, http.StatusOK, s)
	})
}

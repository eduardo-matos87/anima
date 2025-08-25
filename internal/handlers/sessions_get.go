package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

func SessionsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		badRequest(w, "use GET")
		return
	}
	uid := getUserID(r)

	// path: /api/sessions/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		badRequest(w, "missing id")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		badRequest(w, "invalid id")
		return
	}

	s, err := queryOneSession(id, uid)
	if err != nil {
		if errIsNotFound(err) {
			notFound(w)
			return
		}
		badRequest(w, "db error")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

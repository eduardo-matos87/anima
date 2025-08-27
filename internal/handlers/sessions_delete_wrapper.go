package handlers

import (
	"net/http"
)

// Wrapper para DELETE /api/sessions/update/{id}
// Reaproveita o mesmo handler que trata PATCH/DELETE internamente.
func SessionsDelete(w http.ResponseWriter, r *http.Request) {
	if sessionsDB == nil {
		http.Error(w, "sessionsDB not set", http.StatusInternalServerError)
		return
	}
	UpdateDeleteSession(sessionsDB).ServeHTTP(w, r)
}

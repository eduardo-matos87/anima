package handlers

import (
	"net/http"
	"os"
)

// Retorna o user_id para escopo das rotas /api/me/*.
// Regra: 1) Header X-User-ID, 2) env USER_ID, 3) um UUID fixo de desenvolvimento.
func getUserID(r *http.Request) string {
	if v := r.Header.Get("X-User-ID"); v != "" {
		return v
	}
	if v := os.Getenv("USER_ID"); v != "" {
		return v
	}
	return "00000000-0000-0000-0000-000000000001"
}

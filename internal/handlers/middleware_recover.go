package handlers

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recover captura panics e devolve 500 em JSON sem derrubar o servidor.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] %v\n%s", rec, debug.Stack())
				jsonWrite(w, http.StatusInternalServerError, map[string]any{
					"error":      "internal_error",
					"request_id": GetRequestID(r),
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

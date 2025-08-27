package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

type recoverErrorResp struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// Recover intercepta panics nos handlers e retorna 500 em JSON.
// Use assim no main.go: handlers.Recover(mux)
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID := r.Header.Get("X-Request-ID")
				log.Printf("[panic] req_id=%s method=%s path=%s panic=%v\n%s",
					reqID, r.Method, r.URL.Path, rec, debug.Stack())

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				// Mesmo que já tenha escrito algo, tentamos sinalizar erro
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(recoverErrorResp{
					Error:     "internal server error",
					RequestID: reqID,
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
)

func atoiEnv(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

// JSONSafe aplica limite de body e valida Content-Type quando há body.
func JSONSafe(next http.Handler) http.Handler {
	limit := atoiEnv("JSON_BODY_LIMIT_BYTES", 1<<20) // 1MB padrão
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Só métodos com body são verificados
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			// Content-Type
			ct := r.Header.Get("Content-Type")
			if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
				http.Error(w, "unsupported media type (use application/json)", http.StatusUnsupportedMediaType)
				return
			}
			// Limite de tamanho
			if r.Body != nil && limit > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, limit)
			}
		}
		next.ServeHTTP(w, r)
	})
}

package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type ctxKey string

const ctxKeyReqID ctxKey = "req_id"
const ctxKeyUserID ctxKey = "user_id"

// RequestID garante X-Request-ID em toda requisição e guarda no context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			reqID = hex.EncodeToString(b)
		}
		ctx := context.WithValue(r.Context(), ctxKeyReqID, reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(r *http.Request) string {
	if v := r.Context().Value(ctxKeyReqID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Helpers para user_id no context
func SetUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKeyUserID, userID))
}

func GetUserID(r *http.Request) string {
	if v := r.Context().Value(ctxKeyUserID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	// fallback: cabeçalho legado
	return r.Header.Get("X-User-ID")
}

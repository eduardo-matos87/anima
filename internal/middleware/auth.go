package middleware

import (
    "net/http"

    "anima/internal/handlers"
)

// JWTOptional applies optional JWT auth: if a valid token is present, it sets
// the user_id in the request context; otherwise it proceeds without rejecting.
func JWTOptional(next http.Handler) http.Handler { return handlers.OptionalAuth(next) }

// JWTRequired enforces a valid JWT. Requests without a valid token receive 401.
// The user_id is injected into the request context on success.
func JWTRequired(next http.Handler) http.Handler { return handlers.RequireAuth(next) }

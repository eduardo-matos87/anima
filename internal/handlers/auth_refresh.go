package handlers

import (
    "encoding/json"
    "net/http"
    "os"
)

type refreshRequest struct {
    Token string `json:"token"`
}

// AuthRefresh: POST /api/auth/refresh
// Valida o token atual e emite um novo com mesmo sub e novo exp.
func AuthRefresh() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var in refreshRequest
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Token == "" {
            badRequest(w, "invalid json or token missing")
            return
        }

        secret := os.Getenv("JWT_SECRET")
        if secret == "" {
            internalErr(w, errInvalidToken)
            return
        }

        claims, err := verifyJWTHS256(in.Token, secret)
        if err != nil || claims.Sub == "" {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        ttl := atoiEnvInt("JWT_EXP_HOURS", 24)
        token, expIn, err := IssueJWTHS256(claims.Sub, ttl, secret)
        if err != nil {
            internalErr(w, err)
            return
        }

        jsonWrite(w, http.StatusOK, loginResponse{
            AccessToken: token,
            ExpiresIn:   expIn,
            TokenType:   "Bearer",
        })
    }
}

package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "os"
    "strings"
)

type loginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type loginResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int64  `json:"expires_in"`
    TokenType   string `json:"token_type"`
}

// AuthLogin: POST /api/auth/login
// Valida email/senha no Postgres (pgcrypto) e retorna JWT.
func AuthLogin(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var in loginRequest
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            badRequest(w, "invalid json")
            return
        }
        email := strings.TrimSpace(strings.ToLower(in.Email))
        if email == "" || in.Password == "" {
            badRequest(w, "email and password required")
            return
        }

        // Verifica credenciais usando pgcrypto (crypt) no PostgreSQL
        var userID string
        err := db.QueryRow(`
            SELECT id::text
            FROM users
            WHERE email = $1 AND password_hash IS NOT NULL AND password_hash = crypt($2, password_hash)
        `, email, in.Password).Scan(&userID)
        if err == sql.ErrNoRows {
            http.Error(w, "invalid credentials", http.StatusUnauthorized)
            return
        }
        if err != nil {
            internalErr(w, err)
            return
        }

        secret := os.Getenv("JWT_SECRET")
        if secret == "" {
            internalErr(w, errInvalidToken)
            return
        }
        ttl := atoiEnvInt("JWT_EXP_HOURS", 24)
        token, expIn, err := IssueJWTHS256(userID, ttl, secret)
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

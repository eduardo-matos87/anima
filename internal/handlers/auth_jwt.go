package handlers

import (
    "crypto/rand"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

type jwtClaims struct {
	Sub string `json:"sub"`
	Exp *int64 `json:"exp,omitempty"`
	Iat *int64 `json:"iat,omitempty"`
	// adicione campos se precisar
}

var errInvalidToken = errors.New("invalid token")

// OptionalAuth: se houver Bearer JWT válido, popula user_id; senão segue.
func OptionalAuth(next http.Handler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := extractUserIDFromJWT(r, secret)
		if userID != "" {
			r = SetUserID(r, userID)
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth: exige JWT válido; senão 401.
func RequireAuth(next http.Handler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := extractUserIDFromJWT(r, secret)
		if err != nil || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, SetUserID(r, userID))
	})
}

func extractUserIDFromJWT(r *http.Request, secret string) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return "", nil
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if secret == "" {
		return "", errInvalidToken
	}
	claims, err := verifyJWTHS256(token, secret)
	if err != nil {
		return "", err
	}
	return claims.Sub, nil
}

// verifyJWTHS256 valida assinatura e exp (se presente).
func verifyJWTHS256(token, secret string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}
	enc := base64.RawURLEncoding

	// header
	hb, err := enc.DecodeString(parts[0])
	if err != nil {
		return nil, errInvalidToken
	}
	var header map[string]any
	if err := json.Unmarshal(hb, &header); err != nil {
		return nil, errInvalidToken
	}
	if alg, _ := header["alg"].(string); alg != "HS256" {
		return nil, errInvalidToken
	}

	// payload
	pb, err := enc.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}
	var claims jwtClaims
	if err := json.Unmarshal(pb, &claims); err != nil {
		return nil, errInvalidToken
	}
	if claims.Sub == "" {
		return nil, errInvalidToken
	}

	// signature
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		return nil, errInvalidToken
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	mac.Write([]byte("."))
	mac.Write([]byte(parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, errInvalidToken
	}

	// exp
	if claims.Exp != nil {
		now := time.Now().Unix()
		if now > *claims.Exp {
			return nil, errInvalidToken
		}
	}
	return &claims, nil
}

// IssueJWTHS256 emite um JWT assinado (HS256) com sub, iat e exp.
// Retorna token e expiresIn (segundos) calculado a partir de ttlHours.
func IssueJWTHS256(sub string, ttlHours int, secret string) (string, int64, error) {
    if sub == "" || secret == "" || ttlHours <= 0 {
        return "", 0, errInvalidToken
    }
    enc := base64.RawURLEncoding

    header := map[string]any{"alg": "HS256", "typ": "JWT"}
    now := time.Now().Unix()
    exp := time.Now().Add(time.Duration(ttlHours) * time.Hour).Unix()
    // random jti to ensure refreshed tokens differ even within same second
    jti := make([]byte, 12)
    _, _ = rand.Read(jti)
    payload := map[string]any{
        "sub": sub,
        "iat": now,
        "exp": exp,
        "jti": base64.RawURLEncoding.EncodeToString(jti),
    }

    hb, _ := json.Marshal(header)
    pb, _ := json.Marshal(payload)

    seg := enc.EncodeToString(hb) + "." + enc.EncodeToString(pb)
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(seg))
    sig := mac.Sum(nil)

    token := seg + "." + enc.EncodeToString(sig)
    return token, exp - now, nil
}

package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "anima/internal/handlers"
)

func TestProtectedRouteRequiresAuth(t *testing.T) {
    t.Setenv("JWT_SECRET", "testsecret")

    protected := handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    rr := httptest.NewRecorder()
    protected.ServeHTTP(rr, req)

    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", rr.Code)
    }
}

func TestRefreshGeneratesNewToken(t *testing.T) {
    t.Setenv("JWT_SECRET", "testsecret")
    t.Setenv("JWT_EXP_HOURS", "1")

    // Issue a valid token
    token, _, err := handlers.IssueJWTHS256("user-123", 1, os.Getenv("JWT_SECRET"))
    if err != nil {
        t.Fatalf("IssueJWTHS256 error: %v", err)
    }

    // Call refresh endpoint
    h := handlers.AuthRefresh()
    body, _ := json.Marshal(map[string]string{"token": token})

    req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200 from refresh, got %d: %s", rr.Code, rr.Body.String())
    }

    var resp struct{
        AccessToken string `json:"access_token"`
        ExpiresIn   int64  `json:"expires_in"`
        TokenType   string `json:"token_type"`
    }
    if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
        t.Fatalf("invalid json: %v", err)
    }
    if resp.AccessToken == "" {
        t.Fatalf("empty access_token")
    }
    if resp.AccessToken == token {
        t.Fatalf("expected a new token different from the old one")
    }

    // Use the new token on a protected handler and assert user_id is preserved
    protected := handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(handlers.GetUserID(r)))
    }))

    req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
    req2.Header.Set("Authorization", "Bearer "+resp.AccessToken)
    rr2 := httptest.NewRecorder()
    protected.ServeHTTP(rr2, req2)

    if rr2.Code != http.StatusOK {
        t.Fatalf("expected 200 with new token, got %d", rr2.Code)
    }
    if rr2.Body.String() != "user-123" {
        t.Fatalf("expected user_id 'user-123', got '%s'", rr2.Body.String())
    }
}

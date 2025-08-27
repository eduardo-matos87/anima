package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimit aplica token-bucket por "chave" (X-User-ID ou IP).
// maxPerMin = número de requisições por minuto concedidas.
func RateLimit(maxPerMin int) func(http.Handler) http.Handler {
	if maxPerMin <= 0 {
		maxPerMin = 60
	}
	type bucket struct {
		tokens     float64
		lastRefill time.Time
	}
	var (
		mu      sync.Mutex
		buckets = make(map[string]*bucket)
		rate    = float64(maxPerMin) / 60.0 // tokens por segundo
	)

	// limpeza simples (opcional)
	go func() {
		t := time.NewTicker(5 * time.Minute)
		for range t.C {
			mu.Lock()
			now := time.Now()
			for k, b := range buckets {
				if now.Sub(b.lastRefill) > 10*time.Minute {
					delete(buckets, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-User-ID")
			if key == "" {
				key = clientIP(r) // helper já existe em overload.go; ou duplique aqui se preferir.
			}
			if key == "" {
				key = "anon"
			}

			now := time.Now()
			mu.Lock()
			b, ok := buckets[key]
			if !ok {
				b = &bucket{tokens: float64(maxPerMin), lastRefill: now}
				buckets[key] = b
			} else {
				// refil
				elapsed := now.Sub(b.lastRefill).Seconds()
				b.tokens += elapsed * rate
				if b.tokens > float64(maxPerMin) {
					b.tokens = float64(maxPerMin)
				}
				b.lastRefill = now
			}

			if b.tokens >= 1 {
				b.tokens -= 1
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}
			mu.Unlock()

			// 429 Too Many Requests
			w.Header().Set("Retry-After", strconv.Itoa(5))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		})
	}
}

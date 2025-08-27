package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

func New(r rate.Limit, burst int) *Limiter {
	return &Limiter{visitors: make(map[string]*rate.Limiter), rate: r, burst: burst}
}

func (l *Limiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.visitors[ip]
	if !ok {
		lim = rate.NewLimiter(l.rate, l.burst)
		l.visitors[ip] = lim
	}
	return lim
}

func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		lim := l.getLimiter(ip)
		if !lim.Allow() {
			w.Header().Set("Retry-After", time.Second.String())
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

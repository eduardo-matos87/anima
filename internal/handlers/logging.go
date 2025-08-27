package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// WrapLogging adiciona logs JSON line por request (inclui req_id e user_id).
func WrapLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(rec, r)
		entry := map[string]any{
			"ts":         time.Now().Format(time.RFC3339Nano),
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     rec.status,
			"bytes":      rec.bytes,
			"dur_ms":     time.Since(start).Milliseconds(),
			"user_id":    GetUserID(r),
			"req_id":     GetRequestID(r),
			"user_agent": r.UserAgent(),
		}
		b, _ := json.Marshal(entry)
		log.Println(string(b))
	})
}

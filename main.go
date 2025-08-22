package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"anima/internal/handlers"

	_ "github.com/lib/pq"
)

func main() {
	// ===== Config =====
	port := getenv("PORT", "8081")
	dsn := getenv("DATABASE_URL", "postgres://anima:anima@localhost:5432/anima?sslmode=disable")

	// ===== DB =====
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("erro abrindo conexão Postgres: %v", err)
	}
	defer db.Close()

	// ping inicial (não derruba o servidor se falhar)
	if err := pingWithTimeout(db, 3*time.Second); err != nil {
		log.Printf("⚠️ aviso: ping ao Postgres falhou: %v", err)
	}

	// ===== Rotas =====
	mux := http.NewServeMux()

	// Healthcheck
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pingWithTimeout(db, 1*time.Second); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("db: down"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Compat legada
	mux.Handle("/treinos/generate", handlers.GenerateTreino(db))

	// API treinos
	mux.Handle("/api/treinos/generate", handlers.GenerateTreino(db))
	mux.Handle("/api/exercises", handlers.ListExercises(db))
	mux.Handle("/api/treinos", handlers.SaveTreino(db))             // POST
	mux.Handle("/api/treinos/by-key/", handlers.GetTreinoByKey(db)) // GET /api/treinos/by-key/{key}
	mux.Handle("/api/treinos/", handlers.GetTreinoByID(db))         // GET /api/treinos/{id}
	mux.Handle("/api/me/summary", handlers.MeSummaryHandler(db))    // GET

	// API Perfil & Métricas
	mux.Handle("/api/me/profile", handlers.UserProfile(db)) // GET/PUT
	mux.Handle("/api/me/metrics", handlers.UserMetrics(db)) // GET/POST

	// ===== Middlewares =====
	handler := withCORS(mux)

	// ===== Servidor =====
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("API Anima ouvindo em http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func pingWithTimeout(db *sql.DB, d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return db.PingContext(ctx)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

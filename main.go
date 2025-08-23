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

	// Compat legada (antes era sem /api)
	mux.Handle("/treinos/generate", handlers.GenerateTreino(db))

	// ===== API =====

	// Perfil & Métricas
	mux.Handle("/api/me/profile", handlers.UserProfile(db))      // GET/PUT
	mux.Handle("/api/me/metrics", handlers.UserMetrics(db))      // GET/POST
	mux.Handle("/api/me/summary", handlers.MeSummaryHandler(db)) // GET

	// Exercícios (catálogo)
	mux.Handle("/api/exercises", handlers.ListExercises(db)) // GET

	// Treinos
	mux.Handle("/api/treinos/generate", handlers.GenerateTreino(db)) // POST
	// Coleção: GET (lista paginada/buscável) + POST (salvar)
	mux.Handle("/api/treinos", handlers.TreinosCollection(db))
	// by-key é mais específico e não conflita com /api/treinos/
	mux.Handle("/api/treinos/by-key/", handlers.GetTreinoByKey(db)) // GET /api/treinos/by-key/{key}
	// Item: GET por ID + PATCH (coach_notes)
	mux.Handle("/api/treinos/", handlers.TreinosItem(db)) // /api/treinos/{id}

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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

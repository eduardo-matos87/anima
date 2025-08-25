package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"anima/internal/handlers"

	_ "github.com/lib/pq"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// /docs com ReDoc (serve /openapi.yaml de ./docs)
func docsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8"/>
  <title>Anima API Docs</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style> html,body{height:100%} body{margin:0} </style>
</head>
<body>
  <redoc spec-url="/openapi.yaml"></redoc>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// ===== Config =====
	port := getenv("PORT", "8081")
	dsn := getenv("DATABASE_URL", "postgres://anima:anima@localhost:5432/anima?sslmode=disable")

	// ===== DB =====
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("erro abrindo Postgres: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("erro ping Postgres: %v", err)
	}
	log.Println("[anima] conectado ao Postgres")

	// Injeta DB nos novos handlers de histórico
	handlers.SetSessionsDB(db)

	// ===== Mux =====
	mux := http.NewServeMux()

	// Docs
	mux.Handle("/openapi.yaml", http.StripPrefix("/", http.FileServer(http.Dir("./docs"))))
	mux.HandleFunc("/docs", docsHandler)

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"ok":false}`)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, `{"ok":true}`)
	})

	// ===== Catálogo =====
	// GET /api/exercises
	mux.HandleFunc("/api/exercises", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.ListExercises(db).ServeHTTP(w, r)
	})

	// ===== Treinos (coleção) =====
	// GET /api/treinos (listagem)
	// POST /api/treinos (ainda não implementado no repo atual -> 501)
	mux.HandleFunc("/api/treinos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.TreinosCollection(db).ServeHTTP(w, r)
		case http.MethodPost:
			http.Error(w, "not implemented (POST /api/treinos)", http.StatusNotImplemented)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ===== Treinos (item) =====
	// GET /api/treinos/{id}
	// PATCH /api/treinos/{id} (coach_notes ainda não implementado -> 501)
	// /api/treinos/by-key/{key} não encontrado no repo -> 501
	mux.HandleFunc("/api/treinos/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/treinos/by-key/") {
			http.Error(w, "not implemented (/api/treinos/by-key)", http.StatusNotImplemented)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handlers.TreinosItem(db).ServeHTTP(w, r)
		case http.MethodPatch:
			http.Error(w, "not implemented (PATCH /api/treinos/{id})", http.StatusNotImplemented)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ===== Gerador v1.1 =====
	// POST /api/treinos/generate
	mux.HandleFunc("/api/treinos/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.GenerateTreino(db).ServeHTTP(w, r)
	})

	// Progressive Overload — GET /api/suggestions/next-load
	mux.HandleFunc("/api/suggestions/next-load", func(w http.ResponseWriter, r *http.Request) {
		handlers.NextLoad(db).ServeHTTP(w, r)
	})

	// ===== Planner semanal =====
	// GET /api/plan/weekly
	// POST /api/plan/weekly/save
	mux.HandleFunc("/api/plan/weekly", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.PlanWeekly(db).ServeHTTP(w, r)
	})
	mux.HandleFunc("/api/plan/weekly/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.PlanWeeklySave(db).ServeHTTP(w, r)
	})

	// ===== Perfil & Métricas =====
	// Não achei handlers no grep; exponho 501 para não quebrar o contrato.
	mux.HandleFunc("/api/me/profile", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented (/api/me/profile)", http.StatusNotImplemented)
	})
	mux.HandleFunc("/api/me/metrics", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented (/api/me/metrics)", http.StatusNotImplemented)
	})
	mux.HandleFunc("/api/me/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.MeSummaryHandler(db).ServeHTTP(w, r)
	})

	// ===== Histórico: Sessions & Sets (novos) =====
	// /api/sessions        GET (list), POST (create)
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.SessionsList(w, r)
		case http.MethodPost:
			handlers.SessionsCreate(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	// /api/sessions/{id} e subrotas /sets
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.Contains(path, "/sets") {
			switch r.Method {
			case http.MethodGet:
				handlers.SetsList(w, r)
			case http.MethodPost:
				handlers.SetsCreate(w, r)
			case http.MethodPatch:
				handlers.SetsPatch(w, r)
			case http.MethodDelete:
				handlers.SetsDelete(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			handlers.SessionsGet(w, r)
		case http.MethodPatch:
			handlers.SessionsPatch(w, r)
		case http.MethodDelete:
			handlers.SessionsDelete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ===== Server =====
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      withCORS(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[anima] escutando em :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro ListenAndServe: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("[anima] encerrando...")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("erro no shutdown: %v", err)
	}
	_ = db.Close()
	log.Println("[anima] bye")
}

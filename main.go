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

<<<<<<< HEAD
func main() {
	// ===== Config =====
	port := getenv("PORT", "8081")
	dsn := getenv("DATABASE_URL", "postgres://anima:anima@localhost:5432/anima?sslmode=disable")

	// ===== DB =====
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("erro abrindo conexão Postgres: %v", err)
=======
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
	}
	return def
}

<<<<<<< HEAD
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
=======
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
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
<<<<<<< HEAD
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
=======
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Admin-Token, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS")
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
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

	// Injeta DB nos wrappers compat (sessions/sets/overload)
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
	// POST /api/treinos (não implementado ainda -> 501)
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
	// PATCH /api/treinos/{id}
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
			handlers.TreinosUpdate(db).ServeHTTP(w, r)
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

	// ===== Overload (compat legacy) =====
	// GET /api/suggestions/next-load
	mux.HandleFunc("/api/suggestions/next-load", func(w http.ResponseWriter, r *http.Request) {
		handlers.NextLoad(w, r)
	})

	// ===== Overload (GET/POST) com rate limit por ENV =====
	// RATE_LIMIT_OVERLOAD (default 60) — por IP/usuário
	mux.Handle("/api/overload/suggest",
		handlers.RateLimit(handlers.AtoiEnvInt("RATE_LIMIT_OVERLOAD", 60))(
			handlers.OverloadSuggest(db),
		),
	)

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

	// /api/sets/{id}  (PATCH/DELETE)
	mux.HandleFunc("/api/sets/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			handlers.SetsPatch(w, r)
		case http.MethodDelete:
			handlers.SetsDelete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ===== Sessions & Sets (wrappers w,r) =====
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

	// /api/sessions/{id} e subrotas /sets e /update
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// /api/sessions/{id}/sets  (GET/POST)
		if strings.Contains(path, "/sets") {
			switch r.Method {
			case http.MethodGet:
				handlers.SetsList(w, r)
			case http.MethodPost:
				handlers.SetsCreate(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// /api/sessions/update/{id} (PATCH/DELETE)
		if strings.Contains(path, "/update/") {
			switch r.Method {
			case http.MethodPatch:
				handlers.SessionsPatch(w, r)
			case http.MethodDelete:
				handlers.SessionsDelete(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// /api/sessions/{id} (GET)
		switch r.Method {
		case http.MethodGet:
			handlers.SessionsGet(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ===== Admin: Overload =====
	// Refresh MV
	mux.Handle("/api/admin/overload/refresh", handlers.AdminOverloadRefresh(db))
	// Logs (list)
	mux.Handle("/api/admin/overload/logs", handlers.AdminOverloadLogs(db))
	// Stats agregadas
	mux.Handle("/api/admin/overload/stats", handlers.AdminOverloadStats(db))
	// Export CSV
	mux.Handle("/api/admin/overload/export.csv", handlers.AdminOverloadExportCSV(db))

	// PATCH /api/sets/batch  (atualização em lote)
	mux.HandleFunc("/api/sets/batch", handlers.SetsBatch)

	// ===== Perfil & Métricas do usuário =====
	mux.Handle("/api/me/profile", handlers.MeProfile(db))        // GET/PATCH
	mux.Handle("/api/me/metrics", handlers.MeMetrics(db))        // GET
	mux.Handle("/api/me/summary", handlers.MeSummaryHandler(db)) // GET

	// ===== Server =====
	srv := &http.Server{
		Addr: ":" + port,
		Handler: withCORS(
			handlers.RequestID(
				handlers.OptionalAuth( // captura user_id de JWT se presente
					handlers.JSONSafe( // limite de body + valida JSON nos métodos com body
						handlers.WrapLogging(
							handlers.Recover(mux),
						),
					),
				),
			),
		),
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

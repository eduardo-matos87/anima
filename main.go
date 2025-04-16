package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"anima/internal/handlers"
	"database/sql"
	_ "github.com/lib/pq"
)

// Inicializa a conexão PostgreSQL
func setupDatabase() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Banco indisponível: %v", err)
	}

	log.Println("Conectado ao PostgreSQL com sucesso!")
	return db
}

func main() {
	db := setupDatabase()

	router := mux.NewRouter()

	// Rota do Handler para gerar treino
	router.HandleFunc("/gerar-treino", handlers.GerarTreino(db)).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor rodando em :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

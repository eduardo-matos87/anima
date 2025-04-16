package main

import (
	"anima/internal/handlers"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 🔌 Conexão com o banco SQLite
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		log.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close()

	// 🌐 Rota de teste: /ping
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// 📥 Endpoints para treinos
	http.HandleFunc("/treino", handlers.GerarTreino(db))         // GET
	http.HandleFunc("/treino/criar", handlers.CriarTreino(db))     // POST

	// 📋 Endpoint para listar exercícios (já existente)
	http.HandleFunc("/exercicios", handlers.ListarExercicios(db))   // GET

	// 📌 Novos Endpoints para consulta de dados
	http.HandleFunc("/objetivos", handlers.ListarObjetivos(db))       // GET /objetivos
	http.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))     // GET /grupos

	// 🚀 Inicia o servidor na porta 8080
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

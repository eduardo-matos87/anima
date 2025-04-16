package main

import (
	// Importa os handlers da API
	"anima/internal/handlers"

	// Pacotes padrão
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// Driver do SQLite
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 🔌 Conexão com o banco SQLite
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		log.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close() // Garante que o banco será fechado ao final

	// 🌐 Rota de teste (ping)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// 📥 Rota GET para buscar treinos com base em nível e objetivo
	http.HandleFunc("/treino", handlers.GerarTreino(db))

	// 📤 Rota POST para cadastrar um novo treino com exercícios
	http.HandleFunc("/treino/criar", handlers.CriarTreino(db))

	// 🧠 Rota GET para listar exercícios por grupo muscular (ex: /exercicios?grupo=peito)
	http.HandleFunc("/exercicios", handlers.ListarExercicios(db))

	// 🚀 Inicia o servidor na porta 8080
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

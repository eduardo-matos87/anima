// @title Anima API
// @version 1.0
// @description API para gerenciamento de treinos e saúde.
// @termsOfService http://swagger.io/terms/

// @contact.name Eduardo Matos
// @contact.email eduardo@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
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
	// Conecta ao banco SQLite
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		log.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close()

	// Rota de teste: /ping
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// Endpoints de Treino
	// @Summary Busca treinos
	// @Description Retorna treinos com base no nível e objetivo.
	// @Tags Treino
	// @Accept json
	// @Produce json
	// @Param nivel query string true "Nível do treino"
	// @Param objetivo query string true "Objetivo do treino"
	// @Success 200 {object} handlers.RespostaTreino
	// @Failure 500 {object} map[string]string
	// @Router /treino [get]
	http.HandleFunc("/treino", handlers.GerarTreino(db)) // GET

	// Endpoint para criar novo treino
	// @Summary Cria um novo treino
	// @Description Cadastra um novo treino e vincula os exercícios.
	// @Tags Treino
	// @Accept json
	// @Produce json
	// @Param treino body handlers.NovoTreino true "Dados do treino"
	// @Success 201 {object} map[string]interface{}
	// @Failure 400 {object} map[string]string
	// @Failure 500 {object} map[string]string
	// @Router /treino/criar [post]
	http.HandleFunc("/treino/criar", handlers.CriarTreino(db)) // POST

	// Endpoints para exercícios, objetivos e grupos também devem ter suas anotações
	// Exemplo: para listar objetivos
	// @Summary Lista os objetivos
	// @Description Retorna a lista de objetivos cadastrados.
	// @Tags Objetivos
	// @Produce json
	// @Success 200 {array} handlers.Objetivo
	// @Router /objetivos [get]
	http.HandleFunc("/objetivos", handlers.ListarObjetivos(db))

	// Exemplo: para listar grupos musculares
	// @Summary Lista grupos musculares
	// @Description Retorna os grupos musculares cadastrados.
	// @Tags Grupos
	// @Produce json
	// @Success 200 {array} handlers.GrupoMuscular
	// @Router /grupos [get]
	http.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))

	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

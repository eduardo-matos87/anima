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

	_ "github.com/mattn/go-sqlite3"              // Driver do SQLite
	httpSwagger "github.com/swaggo/http-swagger" // Handler para Swagger UI
	_ "anima/docs"                              // Importa a documentação gerada pelo swag
)

func main() {
	// 🔌 Conecta ao banco SQLite (arquivo anima.db na raiz)
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		log.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close()

	// 🌐 Rota de teste: /ping
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// -----------------------------------------------------------------------------
	// Endpoints de Treino
	// -----------------------------------------------------------------------------

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
	http.HandleFunc("/treino", handlers.GerarTreino(db)) // GET para consultar treinos

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
	http.HandleFunc("/treino/criar", handlers.CriarTreino(db)) // POST para criar treino

	// -----------------------------------------------------------------------------
	// Endpoints de Exercícios, Objetivos e Grupos
	// -----------------------------------------------------------------------------
	// (Assumindo que esses handlers foram implementados em outros arquivos)
	http.HandleFunc("/exercicios", handlers.ListarExercicios(db))
	http.HandleFunc("/objetivos", handlers.ListarObjetivos(db))
	http.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))

	// -----------------------------------------------------------------------------
	// Endpoints de Usuário: Registro e Login
	// -----------------------------------------------------------------------------
	// @Summary Registra um novo usuário
	// @Description Cria um usuário com nome, email e senha.
	// @Tags User
	// @Accept json
	// @Produce json
	// @Param user body handlers.Credentials true "Dados de registro"
	// @Success 201 {object} map[string]interface{}
	// @Router /register [post]
	http.HandleFunc("/register", handlers.RegisterUser(db))

	// @Summary Login de usuário
	// @Description Autentica o usuário e retorna um token JWT.
	// @Tags User
	// @Accept json
	// @Produce json
	// @Param credentials body handlers.Credentials true "Dados de login"
	// @Success 200 {object} map[string]string
	// @Router /login [post]
	http.HandleFunc("/login", handlers.LoginUser(db))

	// -----------------------------------------------------------------------------
	// Rota para Documentação Swagger
	// -----------------------------------------------------------------------------
	// Acesse a documentação em: http://localhost:8080/swagger/index.html
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	// 🚀 Inicia o servidor na porta 8080
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

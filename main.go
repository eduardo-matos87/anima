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
	"anima/internal/handlers"              // Pacote com os handlers dos endpoints
	"database/sql"                         // SQL padrão para conexão com o banco
	"fmt"                                  // Formatação e impressão de mensagens
	"log"                                  // Logging de erros e mensagens
	"net/http"                             // Servidor HTTP
	
	_ "github.com/mattn/go-sqlite3"          // Driver SQLite3
	httpSwagger "github.com/swaggo/http-swagger" // Handler para Swagger UI (documentação)
	_ "anima/docs"                         // Importa a documentação gerada pelo swag (Swagger)
)

func main() {
	// 🔌 Conecta ao banco de dados SQLite (arquivo anima.db na raiz do projeto)
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		log.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close() // Garante que a conexão com o banco será fechada quando a função main terminar

	// 🌐 Rota de teste: /ping
	// Serve para verificar se o servidor está ativo
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// -----------------------------------------------------------------------------
	// Endpoints de Treino
	// -----------------------------------------------------------------------------

	// GET /treino
	// @Summary Busca treinos
	// @Description Retorna treinos com base nos parâmetros 'nivel' e 'objetivo'.
	// @Tags Treino
	// @Accept json
	// @Produce json
	// @Param nivel query string true "Nível do treino"
	// @Param objetivo query string true "Objetivo do treino"
	// @Success 200 {object} handlers.RespostaTreino
	// @Router /treino [get]
	http.HandleFunc("/treino", handlers.GerarTreino(db))

	// POST /treino/criar
	// @Summary Cria um novo treino
	// @Description Cadastra um novo treino e vincula os exercícios.
	// @Tags Treino
	// @Accept json
	// @Produce json
	// @Param treino body handlers.NovoTreino true "Dados do novo treino"
	// @Success 201 {object} map[string]interface{}
	// @Router /treino/criar [post]
	// Aqui, estamos utilizando um middleware para proteger a rota com autenticação.
	http.HandleFunc("/treino/criar", handlers.AuthMiddleware(handlers.CriarTreino(db)))

	// -----------------------------------------------------------------------------
	// Endpoints para Consultas de Exercícios, Objetivos e Grupos Musculares
	// -----------------------------------------------------------------------------

	// GET /exercicios?grupo=peito
	// Retorna os exercícios filtrados por grupo muscular
	http.HandleFunc("/exercicios", handlers.ListarExercicios(db))

	// GET /objetivos
	// Retorna a lista de objetivos cadastrados (ex: Emagrecimento, Ganho de massa magra, Resistência física)
	http.HandleFunc("/objetivos", handlers.ListarObjetivos(db))

	// GET /grupos
	// Retorna os grupos musculares cadastrados (ex: Peito, Costas, Pernas, etc.)
	http.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))

	// -----------------------------------------------------------------------------
	// Endpoints de Usuário: Registro e Login
	// -----------------------------------------------------------------------------

	// POST /register
	// @Summary Registra um novo usuário
	// @Description Cria um novo usuário com nome, email e senha.
	// @Tags User
	// @Accept json
	// @Produce json
	// @Param user body handlers.Credentials true "Dados de registro do usuário"
	// @Success 201 {object} map[string]interface{}
	// @Router /register [post]
	http.HandleFunc("/register", handlers.RegisterUser(db))

	// POST /login
	// @Summary Efetua login do usuário
	// @Description Autentica o usuário e retorna um token JWT.
	// @Tags User
	// @Accept json
	// @Produce json
	// @Param credentials body handlers.Credentials true "Dados de login do usuário"
	// @Success 200 {object} map[string]string
	// @Router /login [post]
	http.HandleFunc("/login", handlers.LoginUser(db))

	// -----------------------------------------------------------------------------
	// Rota para Documentação Swagger
	// -----------------------------------------------------------------------------
	// Serve a documentação da API gerada pelo swag em:
	// http://localhost:8080/swagger/index.html
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	// 🚀 Inicia o servidor na porta 8080
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

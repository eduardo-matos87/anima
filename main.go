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
	"anima/internal/handlers"            // Pacote com os handlers (funções de endpoint)
	"database/sql"                       // Manipulação de banco de dados SQL
	"fmt"                                // Formatação e impressão de strings
	"log"                                // Log padrão para erros fatais
	"net/http"                           // Servidor HTTP
	"os"                                 // Interação com o sistema operacional

	// Importa o driver SQLite3 para manipulação do banco de dados
	_ "github.com/mattn/go-sqlite3"
	
	// Importa o Swagger UI para servir a documentação
	httpSwagger "github.com/swaggo/http-swagger"
	_ "anima/docs"                       // Importa a documentação gerada pelo swag

	// Logrus para logging avançado
	logrus "github.com/sirupsen/logrus"
)

func main() {
	// ----------------------------------------------------------------------------
	// Configuração do Logrus para registro de logs
	// ----------------------------------------------------------------------------
	// Define o formato para incluir timestamps completos.
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	
	// Define o caminho do arquivo de log. Certifique-se de ter permissões para escrever em /var/log.
	logFilePath := "/var/log/anima.log"
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Se não for possível abrir o arquivo, registra a informação e utiliza o stderr.
		logrus.Info("Falha ao abrir /var/log/anima.log, utilizando stderr: ", err)
	} else {
		// Redireciona os logs para o arquivo aberto.
		logrus.SetOutput(file)
	}

	// ----------------------------------------------------------------------------
	// Conexão com o Banco de Dados SQLite
	// ----------------------------------------------------------------------------
	// Abre o arquivo de banco de dados "anima.db", que deve estar na raiz do projeto.
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		logrus.Fatal("Erro ao conectar no banco de dados:", err)
	}
	// Garante que a conexão será fechada quando o main terminar.
	defer db.Close()

	// ----------------------------------------------------------------------------
	// Configuração dos Endpoints (Rotas) da API
	// ----------------------------------------------------------------------------

	// Rota de teste para verificar se o servidor está ativo.
	// Quando acessado, retorna "pong".
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// ----------------------------
	// Endpoints de Treino
	// ----------------------------

	// GET /treino:
	// Busca e retorna treinos com base no "nivel" e "objetivo" passados como query string.
	http.HandleFunc("/treino", handlers.GerarTreino(db))

	// POST /treino/criar:
	// Cria um novo treino e vincula os exercícios.
	// Esse endpoint é protegido; utiliza o middleware de autenticação.
	http.HandleFunc("/treino/criar", handlers.AuthMiddleware(handlers.CriarTreino(db)))

	// ----------------------------
	// Endpoints para Consultas de Dados
	// ----------------------------

	// GET /exercicios:
	// Lista os exercícios filtrados por grupo muscular (ex: /exercicios?grupo=peito).
	http.HandleFunc("/exercicios", handlers.ListarExercicios(db))

	// GET /objetivos:
	// Retorna a lista de objetivos cadastrados (ex: Emagrecimento, Ganho de massa magra, Resistência física).
	http.HandleFunc("/objetivos", handlers.ListarObjetivos(db))

	// GET /grupos:
	// Retorna a lista de grupos musculares cadastrados (ex: Peito, Costas, Pernas, etc.).
	http.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))

	// ----------------------------
	// Endpoints de Usuário: Registro e Login
	// ----------------------------

	// POST /register:
	// Registra um novo usuário com nome, email e senha.
	http.HandleFunc("/register", handlers.RegisterUser(db))

	// POST /login:
	// Autentica o usuário e retorna um token JWT se os dados estiverem corretos.
	http.HandleFunc("/login", handlers.LoginUser(db))

	// ----------------------------
	// Rota para Documentação Swagger
	// ----------------------------
	// A documentação gerada pelo swag ficará disponível em:
	// http://localhost:8080/swagger/index.html
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	// ----------------------------------------------------------------------------
	// Inicialização do Servidor HTTP
	// ----------------------------------------------------------------------------
	// Exibe uma mensagem de inicialização e inicia o servidor na porta 8080.
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

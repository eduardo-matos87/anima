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
	"anima/internal/handlers"             // Pacote com os handlers (rotas) da API
	"database/sql"                        // Pacote para conexão e manipulação de banco de dados SQL
	"fmt"                                 // Pacote para formatação de strings e impressão
	"log"                                 // Pacote padrão para log (usado além do Logrus para tratamento fatal)
	"net/http"                            // Pacote para criação do servidor HTTP
	"os"                                  // Pacote para interação com o sistema operacional

	// Importa o driver SQLite3 para conexão com o banco
	_ "github.com/mattn/go-sqlite3"

	// Swagger UI para servir a documentação gerada
	httpSwagger "github.com/swaggo/http-swagger"
	_ "anima/docs"                        // Importa a documentação gerada pelo swag (Swagger)

	// Logrus para log avançado e customizado
	logrus "github.com/sirupsen/logrus"
)

func main() {
	// -----------------------------------------------------------------------------
	// Configuração do Logrus para registro de logs
	// -----------------------------------------------------------------------------
	// Define o formato dos logs para incluir o timestamp completo
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// Tenta abrir (ou criar) o arquivo de log "anima.log" para persistência de logs
	file, err := os.OpenFile("anima.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// Se aberto com sucesso, redireciona as saídas de log para esse arquivo
		logrus.SetOutput(file)
	} else {
		// Caso contrário, registra um aviso e utiliza o stderr padrão
		logrus.Info("Falha ao abrir o arquivo de log, usando stderr")
	}

	// -----------------------------------------------------------------------------
	// Conexão com o Banco de Dados SQLite
	// -----------------------------------------------------------------------------
	// Abre a conexão com o arquivo de banco de dados "anima.db" que deve estar na raiz do projeto
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		logrus.Fatal("Erro ao conectar no banco de dados:", err)
	}
	// Garante que a conexão será fechada ao término da execução do programa
	defer db.Close()

	// -----------------------------------------------------------------------------
	// Configuração dos Endpoints (Rotas) da API
	// -----------------------------------------------------------------------------

	// Rota de teste para verificar se o servidor está ativo.
	// Acessando /ping, o servidor responde com "pong".
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// ----------------------------
	// Endpoints de Treino
	// ----------------------------

	// GET /treino
	// Busca treinos com base no "nivel" e "objetivo" passados via query string.
	http.HandleFunc("/treino", handlers.GerarTreino(db))

	// POST /treino/criar
	// Cria um novo treino e relaciona com os exercícios fornecidos.
	// Essa rota é protegida: somente será executada se o token JWT enviado no header for válido.
	http.HandleFunc("/treino/criar", handlers.AuthMiddleware(handlers.CriarTreino(db)))

	// ----------------------------
	// Endpoints de Exercícios, Objetivos e Grupos Musculares
	// ----------------------------

	// GET /exercicios
	// Lista os exercícios filtrados, de acordo com o grupo muscular (ex.: /exercicios?grupo=peito)
	http.HandleFunc("/exercicios", handlers.ListarExercicios(db))

	// GET /objetivos
	// Retorna a lista de objetivos cadastrados no banco (ex.: Emagrecimento, Ganho de massa magra, Resistência física)
	http.HandleFunc("/objetivos", handlers.ListarObjetivos(db))

	// GET /grupos
	// Retorna a lista de grupos musculares (ex.: Peito, Costas, Pernas, etc.)
	http.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))

	// ----------------------------
	// Endpoints de Usuário: Registro e Login
	// ----------------------------

	// POST /register
	// Registra um novo usuário fornecendo nome, email e senha.
	http.HandleFunc("/register", handlers.RegisterUser(db))

	// POST /login
	// Autentica o usuário e, se os dados estiverem corretos, retorna um token JWT.
	http.HandleFunc("/login", handlers.LoginUser(db))

	// (Opcional) Você pode adicionar um endpoint para ver os logs da aplicação
	// Exemplo: http.HandleFunc("/logs", handlers.LogsHandler())

	// ----------------------------
	// Endpoint para Documentação Swagger
	// ----------------------------
	// A documentação gerada pelo swag ficará disponível em:
	// http://localhost:8080/swagger/index.html
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	// Logs 
	http.HandleFunc("/register", handlers.RegisterUser(db))
	http.HandleFunc("/login", handlers.LoginUser(db))
	
	// -----------------------------------------------------------------------------
	// Inicialização do Servidor HTTP
	// -----------------------------------------------------------------------------
	// Exibe uma mensagem informativa no console e inicia o servidor na porta 8080.
	fmt.Println("Servidor rodando em http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		// Se ocorrer um erro durante a inicialização do servidor, registra o erro e encerra a aplicação.
		logrus.Fatal("Erro ao iniciar o servidor:", err)
	}
}

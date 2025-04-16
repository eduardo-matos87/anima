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
	"anima/internal/handlers"            // Handlers da API
	"database/sql"                       // Conexão SQL
	"fmt"                                // Impressão formatada
	"log"                                // Fail‑fast em ListenAndServe
	"net/http"                           // HTTP server
	"os"                                 // Acesso a arquivos do SO

	_ "github.com/mattn/go-sqlite3"      // Driver SQLite3
	httpSwagger "github.com/swaggo/http-swagger"
	_ "anima/docs"                       // Documentação Swagger

	logrus "github.com/sirupsen/logrus"  // Logging avançado
)

func main() {
	// ----------------------------------------------------------------------------
	// 1) CONFIGURAÇÃO DE LOG (modo DEBUG)
	// ----------------------------------------------------------------------------
	// Use timestamps completos e nível DEBUG para ver todos os logs.
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)

	// Tenta abrir /var/log/anima.log para persistir logs.
	const logFilePath = "/var/log/anima.log"
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Se falhar, avisa e segue logging no stderr
		logrus.Warnf("Não foi possível abrir %s, usando stderr: %v", logFilePath, err)
	} else {
		logrus.Infof("Registrando logs em %s", logFilePath)
		logrus.SetOutput(f)
	}

	// ----------------------------------------------------------------------------
	// 2) CONEXÃO COM O BANCO DE DADOS
	// ----------------------------------------------------------------------------
	// Abre o SQLite (arquivo anima.db na raiz)
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		logrus.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close() // Fecha conexão ao final

	// ----------------------------------------------------------------------------
	// 3) CONFIGURAÇÃO DE ROTAS / ENDPOINTS
	// ----------------------------------------------------------------------------

	// Rota de health‑check
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// ----------------------------
	// Endpoints de Treino
	// ----------------------------

	// GET /treino → busca treinos por nivel e objetivo
	http.HandleFunc("/treino", handlers.GerarTreino(db))

	// POST /treino/criar → cria novo treino (protegido por JWT)
	http.HandleFunc("/treino/criar", handlers.AuthMiddleware(handlers.CriarTreino(db)))

	// ----------------------------
	// Endpoints de Consulta
	// ----------------------------

	// GET /exercicios → lista exercícios por grupo (query param ?grupo=)
	http.HandleFunc("/exercicios", handlers.ListarExercicios(db))

	// GET /objetivos → lista objetivos
	http.HandleFunc("/objetivos", handlers.ListarObjetivos(db))

	// GET /grupos → lista grupos musculares
	http.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))

	// ----------------------------
	// Endpoints de Usuário
	// ----------------------------

	// POST /register → registra novo usuário
	http.HandleFunc("/register", handlers.RegisterUser(db))

	// POST /login → autentica usuário e retorna JWT
	http.HandleFunc("/login", handlers.LoginUser(db))

	// ----------------------------
	// Swagger UI (documentação)
	// ----------------------------
	// Acesse em: http://localhost:8080/swagger/index.html
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	// ----------------------------------------------------------------------------
	// 4) INICIALIZAÇÃO DO SERVIDOR HTTP
	// ----------------------------------------------------------------------------
	logrus.Info("Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logrus.Fatal("Erro no ListenAndServe:", err)
	}
}

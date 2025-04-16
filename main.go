// Arquivo: anima/main.go

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
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"anima/internal/handlers"
	_ "anima/docs"                    // Documentação Swagger gerada pelo swag
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/mattn/go-sqlite3"   // Driver SQLite3
	logrus "github.com/sirupsen/logrus"
)

// corsMiddleware adiciona cabeçalhos CORS para permitir chamadas do front‑end
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// DEBUG: mostra a origem da requisição
		logrus.Debugf("CORS %s request from %s", r.Method, r.Header.Get("Origin"))

		// Permite todas as origens (use '*' apenas em dev)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Métodos permitidos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Cabeçalhos permitidos
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Se for preflight, responde OK imediatamente
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Caso contrário, chama o próximo handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// ----------------------------------------------------------------------------
	// 1) CONFIGURAÇÃO DE LOG (modo DEBUG)
	// ----------------------------------------------------------------------------
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)

	const logFile = "/var/log/anima.log"
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Warnf("Não foi possível abrir %s, usando stderr: %v", logFile, err)
	} else {
		logrus.Infof("Registrando logs em %s", logFile)
		logrus.SetOutput(f)
	}

	// ----------------------------------------------------------------------------
	// 2) CONEXÃO COM O BANCO DE DADOS
	// ----------------------------------------------------------------------------
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		logrus.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close()

	// ----------------------------------------------------------------------------
	// 3) CONFIGURAÇÃO DE ROTAS / ENDPOINTS
	// ----------------------------------------------------------------------------
	mux := http.NewServeMux()

	// Health‑check (GET /ping)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// Treinos
	// GET  /treino         → Gera treino pontual
	mux.HandleFunc("/treino", handlers.GerarTreino(db))
	// POST /treino/criar   → Cria novo treino (protegido por JWT)
	mux.HandleFunc("/treino/criar", handlers.AuthMiddleware(handlers.CriarTreino(db)))
	// GET  /treinos        → Lista todos os treinos cadastrados (protegido)
	mux.HandleFunc("/treinos", handlers.AuthMiddleware(handlers.ListarTreinos(db)))

	// Consultas auxiliares
	mux.HandleFunc("/exercicios", handlers.ListarExercicios(db))         // GET /exercicios?grupo=
	mux.HandleFunc("/objetivos", handlers.ListarObjetivos(db))           // GET /objetivos
	mux.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))       // GET /grupos

	// Usuário
	mux.HandleFunc("/register", handlers.RegisterUser(db))               // POST /register
	mux.HandleFunc("/login", handlers.LoginUser(db))                     // POST /login

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// ----------------------------------------------------------------------------
	// 4) INICIALIZAÇÃO DO SERVIDOR COM CORS
	// ----------------------------------------------------------------------------
	handler := corsMiddleware(mux)

	logrus.Info("Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		logrus.Fatal("Erro ao iniciar servidor:", err)
	}
}

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
	"anima/internal/handlers"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "anima/docs"

	logrus "github.com/sirupsen/logrus"
)

// corsMiddleware envolve um handler e adiciona os headers CORS,
// além de responder as requisições OPTIONS.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permite a origem do front-end (ajuste se mudar de porta ou domínio)
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		// Métodos permitidos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Headers que o cliente pode enviar
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Se for preflight, encerra aqui
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Caso contrário, segue para o handler real
		next.ServeHTTP(w, r)
	})
}

func main() {
	// ----------------------------------------------------------------------------
	// Logrus em modo DEBUG
	// ----------------------------------------------------------------------------
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)

	const logFilePath = "/var/log/anima.log"
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Warnf("Não foi possível abrir %s, usando stderr: %v", logFilePath, err)
	} else {
		logrus.Infof("Registrando logs em %s", logFilePath)
		logrus.SetOutput(f)
	}

	// ----------------------------------------------------------------------------
	// Banco de dados
	// ----------------------------------------------------------------------------
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		logrus.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close()

	// ----------------------------------------------------------------------------
	// Registro de rotas
	// ----------------------------------------------------------------------------
	mux := http.NewServeMux()

	// Health‑check
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	// Treinos
	mux.HandleFunc("/treino", handlers.GerarTreino(db))
	mux.HandleFunc("/treino/criar", handlers.AuthMiddleware(handlers.CriarTreino(db)))

	// Consultas
	mux.HandleFunc("/exercicios", handlers.ListarExercicios(db))
	mux.HandleFunc("/objetivos", handlers.ListarObjetivos(db))
	mux.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))

	// Usuário
	mux.HandleFunc("/register", handlers.RegisterUser(db))
	mux.HandleFunc("/login", handlers.LoginUser(db))

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// ----------------------------------------------------------------------------
	// Inicia o servidor com CORS habilitado
	// ----------------------------------------------------------------------------
	handlerWithCORS := corsMiddleware(mux)

	logrus.Info("Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		logrus.Fatal("Erro no ListenAndServe:", err)
	}
}

// Arquivo: main.go

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

	"anima/internal/handlers"                   // Handlers da aplicação
	_ "anima/docs"                              // Swagger docs
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/mattn/go-sqlite3"             // Driver SQLite3
	logrus "github.com/sirupsen/logrus"         // Log avançado
)

// corsMiddleware adiciona headers CORS e responde OPTIONS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Debugf("CORS %s from %s", r.Method, r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1) Log em modo DEBUG
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)

	const logFile = "/var/log/anima.log"
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Warnf("Não foi possível abrir %s: %v", logFile, err)
	} else {
		logrus.Infof("Registrando logs em %s", logFile)
		logrus.SetOutput(f)
	}

	// 2) Conecta ao banco SQLite
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		logrus.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close()

	// 3) Configuração de rotas
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})
	mux.HandleFunc("/treino", handlers.GerarTreino(db))
	mux.HandleFunc("/treino/criar", handlers.AuthMiddleware(handlers.CriarTreino(db)))
	mux.HandleFunc("/treinos", handlers.AuthMiddleware(handlers.ListarTreinos(db)))
	mux.HandleFunc("/exercicios", handlers.ListarExercicios(db))
	mux.HandleFunc("/objetivos", handlers.ListarObjetivos(db))
	mux.HandleFunc("/grupos", handlers.ListarGruposMusculares(db))
	mux.HandleFunc("/register", handlers.RegisterUser(db))
	mux.HandleFunc("/login", handlers.LoginUser(db))
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// 4) Inicia servidor com CORS
	handler := corsMiddleware(mux)
	logrus.Info("Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		logrus.Fatal("Erro ao iniciar servidor:", err)
	}
}

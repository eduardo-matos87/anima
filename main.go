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

// corsMiddleware adiciona os headers CORS e responde ao OPTIONS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// DEBUG: de onde veio a requisição?
		origin := r.Header.Get("Origin")
		logrus.Debugf("CORS preflight/origin: %s %s", r.Method, origin)

		// Para dev, libera todas as origens:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Se quiser restringir apenas ao React, use:
		// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Se for preflight OPTIONS, retorna OK imediatamente
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Caso contrário, siga para o handler real
		next.ServeHTTP(w, r)
	})
}

func main() {
	// --------------------------------------------------------------------------
	// CONFIGURAÇÃO DE LOG (DEBUG)
	// --------------------------------------------------------------------------
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)

	const logFile = "/var/log/anima.log"
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Warnf("não foi possível abrir %s, usando stderr: %v", logFile, err)
	} else {
		logrus.Infof("registrando logs em %s", logFile)
		logrus.SetOutput(f)
	}

	// --------------------------------------------------------------------------
	// CONEXÃO COM O BANCO DE DADOS
	// --------------------------------------------------------------------------
	db, err := sql.Open("sqlite3", "./anima.db")
	if err != nil {
		logrus.Fatal("Erro ao conectar no banco de dados:", err)
	}
	defer db.Close()

	// --------------------------------------------------------------------------
	// REGISTRO DE ROTAS
	// --------------------------------------------------------------------------
	mux := http.NewServeMux()

	// Health-check
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

	mux.HandleFunc("/treinos", handlers.AuthMiddleware(handlers.ListarTreinos(db)))

	// --------------------------------------------------------------------------
	// START SERVER COM CORS
	// --------------------------------------------------------------------------
	handler := corsMiddleware(mux)
	logrus.Info("Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		logrus.Fatal("ListenAndServe:", err)
	}
}

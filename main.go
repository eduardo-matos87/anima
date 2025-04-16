package main

import (
	"fmt"
	"log"
	"net/http"
	"anima/internal/handlers"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	http.HandleFunc("/treino", handlers.GerarTreino)

	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

http.HandleFunc("/treino", handlers.GerarTreino(db))         // GET
http.HandleFunc("/treino/criar", handlers.CriarTreino(db))   // POST
http.HandleFunc("/exercicios", handlers.ListarExercicios(db))


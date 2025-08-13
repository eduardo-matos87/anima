package main

import (
	"log"
	"net/http"
	"time"

	appdb "anima/internal/db"
	"anima/internal/handlers"
	httpmw "anima/internal/http"
)

func main() {
	// conecta no Postgres
	pg := appdb.Connect()
	defer pg.Close()

	// instancia o handler
	gen := handlers.NewGenerateHandler(pg)

	// define rotas
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/generate", gen.GenerateWorkout)

	// aplica CORS
	handler := httpmw.CORS(mux)

	// configura e inicia servidor
	srv := &http.Server{
		Addr:              ":8081", // backend em 8081
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("API Anima ouvindo em http://localhost:8081")
	log.Fatal(srv.ListenAndServe())
}

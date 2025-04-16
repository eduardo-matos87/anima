// Arquivo: anima/internal/handlers/treino.go

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// RespostaTreino é o JSON de resposta do GET /treino
type RespostaTreino struct {
	Dia        string   `json:"dia"`
	Exercicios []string `json:"exercicios"`
}

// GerarTreino retorna um HandlerFunc para o GET /treino.
// Ele lê os query params "nivel" e "objetivo" e gera um treino dummy em JSON.
func GerarTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Aqui você pode ler r.URL.Query().Get("nivel") e "objetivo"
		// e gerar dinamicamente. Exemplo fixo:
		resp := RespostaTreino{
			Dia:        "Segunda",
			Exercicios: []string{"Supino reto", "Supino inclinado", "Crucifixo", "Tríceps testa"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

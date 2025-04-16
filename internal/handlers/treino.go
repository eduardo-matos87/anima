package handlers

import (
  "database/sql"
  "encoding/json"
  "net/http"
)

// RespostaTreino estrutura a resposta do GET /treino
type RespostaTreino struct {
  Dia        string   `json:"dia"`
  Exercicios []string `json:"exercicios"`
}

// GerarTreino retorna um HandlerFunc para GET /treino
func GerarTreino(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    resp := RespostaTreino{
      Dia:        "Segunda",
      Exercicios: []string{"Supino reto", "Supino inclinado", "Crucifixo", "Tríceps testa"},
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
  }
}

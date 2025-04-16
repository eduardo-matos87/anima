// Arquivo: internal/handlers/treinos.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// TreinoDetail para GET /treinos
type TreinoDetail struct {
	ID         int64    `json:"id"`
	Nivel      string   `json:"nivel"`
	Objetivo   string   `json:"objetivo"`
	Dias       int      `json:"dias"`
	Divisao    string   `json:"divisao"`
	Exercicios []string `json:"exercicios"`
}

// ListarTreinos retorna todos os treinos com exercícios
func ListarTreinos(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id,nivel,objetivo,dias,divisao FROM treinos")
		if err != nil {
			http.Error(w, "Erro ao buscar treinos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var out []TreinoDetail
		for rows.Next() {
			var t TreinoDetail
			rows.Scan(&t.ID, &t.Nivel, &t.Objetivo, &t.Dias, &t.Divisao)
			exRows, _ := db.Query(`
				SELECT e.nome 
				FROM treino_exercicios te 
				JOIN exercicios e ON e.id=te.exercicio_id 
				WHERE te.treino_id=?`, t.ID)
			for exRows.Next() {
				var nome string
				exRows.Scan(&nome)
				t.Exercicios = append(t.Exercicios, nome)
			}
			exRows.Close()
			out = append(out, t)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}

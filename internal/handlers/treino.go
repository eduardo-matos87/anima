// Arquivo: internal/handlers/treinos.go

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// TreinoDetail representa um treino com seus exercícios
type TreinoDetail struct {
	ID         int64    `json:"id"`
	Nivel      string   `json:"nivel"`
	Objetivo   string   `json:"objetivo"`
	Dias       int      `json:"dias"`
	Divisao    string   `json:"divisao"`
	Exercicios []string `json:"exercicios"`
}

// ListarTreinos retorna todos os treinos cadastrados, protegidos por JWT
func ListarTreinos(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) Busca todos os treinos
		rows, err := db.Query("SELECT id, nivel, objetivo, dias, divisao FROM treinos")
		if err != nil {
			http.Error(w, "Erro ao buscar treinos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var treinos []TreinoDetail

		for rows.Next() {
			var t TreinoDetail
			if err := rows.Scan(&t.ID, &t.Nivel, &t.Objetivo, &t.Dias, &t.Divisao); err != nil {
				http.Error(w, "Erro ao ler treinos", http.StatusInternalServerError)
				return
			}
			// 2) Para cada treino, busca os nomes dos exercícios relacionados
			exRows, err := db.Query(`
				SELECT e.nome 
				FROM treino_exercicios te 
				JOIN exercicios e ON e.id = te.exercicio_id 
				WHERE te.treino_id = ?`, t.ID)
			if err != nil {
				http.Error(w, "Erro ao buscar exercícios", http.StatusInternalServerError)
				return
			}
			for exRows.Next() {
				var nome string
				exRows.Scan(&nome)
				t.Exercicios = append(t.Exercicios, nome)
			}
			exRows.Close()

			treinos = append(treinos, t)
		}

		// 3) Retorna como JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(treinos)
	}
}

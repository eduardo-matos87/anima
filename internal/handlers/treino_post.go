// Arquivo: internal/handlers/treino_post.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// NovoTreino payload de criação
type NovoTreino struct {
	Nivel      string `json:"nivel"`
	Objetivo   string `json:"objetivo"`
	Dias       int    `json:"dias"`
	Divisao    string `json:"divisao"`
	Exercicios []int  `json:"exercicios"`
}

// CriarTreino insere um treino e seus exercícios
func CriarTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var nt NovoTreino
		if err := json.NewDecoder(r.Body).Decode(&nt); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		tx, _ := db.Begin()
		res, err := tx.Exec(
			"INSERT INTO treinos(nivel,objetivo,dias,divisao) VALUES(?,?,?,?)",
			nt.Nivel, nt.Objetivo, nt.Dias, nt.Divisao,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Erro ao salvar treino", http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		stmt, _ := tx.Prepare("INSERT INTO treino_exercicios(treino_id,exercicio_id) VALUES(?,?)")
		for _, exID := range nt.Exercicios {
			stmt.Exec(id, exID)
		}
		tx.Commit()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem":  "Treino criado com sucesso",
			"treino_id": id,
		})
	}
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// NovoTreino representa o corpo da requisição para criar um novo treino
type NovoTreino struct {
	Nivel      string  `json:"nivel"`       // Ex: iniciante, intermediario
	Objetivo   string  `json:"objetivo"`    // Ex: emagrecimento, hipertrofia
	Dias       int     `json:"dias"`        // Ex: 3, 5
	Divisao    string  `json:"divisao"`     // Ex: A, B, C
	Exercicios []int64 `json:"exercicios"`  // IDs dos exercícios vinculados
}

// CriarTreino cadastra um novo treino no banco e relaciona com exercícios
func CriarTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var novoTreino NovoTreino

		// 🔎 Decodifica o JSON enviado na requisição
		if err := json.NewDecoder(r.Body).Decode(&novoTreino); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// 🔐 Inicia uma transação para garantir integridade
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
			return
		}

		// 🧱 Insere o novo treino na tabela 'treinos'
		res, err := tx.Exec(`
			INSERT INTO treinos (nivel, objetivo, dias, divisao)
			VALUES (?, ?, ?, ?)`,
			novoTreino.Nivel, novoTreino.Objetivo, novoTreino.Dias, novoTreino.Divisao,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Erro ao salvar treino", http.StatusInternalServerError)
			return
		}

		// 🔄 Recupera o ID do treino recém-criado
		treinoID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			http.Error(w, "Erro ao obter ID do treino", http.StatusInternalServerError)
			return
		}

		// 🔗 Relaciona os exercícios com o treino
		for _, exercicioID := range novoTreino.Exercicios {
			_, err := tx.Exec(`
				INSERT INTO treino_exercicios (treino_id, exercicio_id)
				VALUES (?, ?)`, treinoID, exercicioID)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Erro ao vincular exercício", http.StatusInternalServerError)
				return
			}
		}

		// ✅ Finaliza a transação
		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao salvar treino", http.StatusInternalServerError)
			return
		}

		// 📤 Envia resposta de sucesso com o ID do treino
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"mensagem":  "Treino criado com sucesso",
			"treino_id": treinoID,
		})
	}
}

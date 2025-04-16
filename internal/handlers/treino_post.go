package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// NovoTreino representa o JSON que será enviado via POST
type NovoTreino struct {
	Nivel      string  `json:"nivel"`       // Ex: iniciante, intermediario
	Objetivo   string  `json:"objetivo"`    // Ex: emagrecimento, hipertrofia
	Dias       int     `json:"dias"`        // Ex: 3, 5
	Divisao    string  `json:"divisao"`     // Ex: A, B, C
	Exercicios []int64 `json:"exercicios"`  // Lista de IDs de exercícios
}

// CriarTreino cadastra um novo treino no banco e relaciona os exercícios
func CriarTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var novoTreino NovoTreino

		// 🔎 Decodifica o JSON da requisição
		if err := json.NewDecoder(r.Body).Decode(&novoTreino); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// 🔐 Inicia transação para garantir atomicidade
		tx, err := db.Begin()
		if err != nil {
			log.Println("Erro ao iniciar transação:", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		// 🧱 Insere treino na tabela principal
		res, err := tx.Exec(`
			INSERT INTO treinos (nivel, objetivo, dias, divisao)
			VALUES (?, ?, ?, ?)`,
			novoTreino.Nivel, novoTreino.Objetivo, novoTreino.Dias, novoTreino.Divisao,
		)
		if err != nil {
			log.Println("Erro INSERT treino:", err)
			tx.Rollback()
			http.Error(w, "Erro ao salvar treino", http.StatusInternalServerError)
			return
		}

		// 🔄 Recupera o ID do treino recém-inserido
		treinoID, err := res.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter ID do treino:", err)
			tx.Rollback()
			http.Error(w, "Erro ao salvar treino", http.StatusInternalServerError)
			return
		}

		// 🔗 Insere os exercícios relacionados ao treino
		for _, exercicioID := range novoTreino.Exercicios {
			_, err := tx.Exec(`
				INSERT INTO treino_exercicios (treino_id, exercicio_id)
				VALUES (?, ?)`, treinoID, exercicioID)
			if err != nil {
				log.Println("Erro ao vincular exercício:", err)
				tx.Rollback()
				http.Error(w, "Erro ao vincular exercício", http.StatusInternalServerError)
				return
			}
		}

		// ✅ Finaliza a transação
		if err := tx.Commit(); err != nil {
			log.Println("Erro ao finalizar transação:", err)
			http.Error(w, "Erro ao salvar treino", http.StatusInternalServerError)
			return
		}

		// 📨 Retorna JSON de sucesso
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"mensagem":  "Treino criado com sucesso",
			"treino_id": treinoID,
		})
	}
}

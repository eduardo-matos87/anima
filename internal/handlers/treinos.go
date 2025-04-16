package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// TreinoSuggestion representa um treino sugerido pelo objetivo
type TreinoSuggestion struct {
	ID         int64    `json:"id"`
	Nivel      string   `json:"nivel"`
	Objetivo   string   `json:"objetivo"`
	Dias       int      `json:"dias"`
	Divisao    string   `json:"divisao"`
	Exercicios []string `json:"exercicios"`
}

// GerarTreino sugere treinos baseado no parâmetro "objetivo" passado na query string
func GerarTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		objetivo := r.URL.Query().Get("objetivo")
		if objetivo == "" {
			http.Error(w, "Parâmetro 'objetivo' é obrigatório", http.StatusBadRequest)
			return
		}

		// Busca treinos que tenham o objetivo solicitado
		rows, err := db.Query(
			"SELECT id, nivel, dias, divisao FROM treinos WHERE objetivo = ?",
			objetivo,
		)
		if err != nil {
			http.Error(w, "Erro ao buscar treinos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sugestoes []TreinoSuggestion
		for rows.Next() {
			var t TreinoSuggestion
			t.Objetivo = objetivo
			if err := rows.Scan(&t.ID, &t.Nivel, &t.Dias, &t.Divisao); err != nil {
				http.Error(w, "Erro ao ler dados de treinos", http.StatusInternalServerError)
				return
			}

			// Busca exercícios relacionados a este treino
			exRows, err := db.Query(
				"SELECT e.nome FROM treino_exercicios te JOIN exercicios e ON e.id = te.exercicio_id WHERE te.treino_id = ?",
				t.ID,
			)
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

			sugestoes = append(sugestoes, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sugestoes)
	}
}

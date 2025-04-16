package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Estrutura do treino sugerido
type Treino struct {
	ID         int      `json:"id"`
	Objetivo   string   `json:"objetivo"`
	Nivel      string   `json:"nivel"`
	Exercicios []string `json:"exercicios"`
}

// GerarTreino gera treino personalizado baseado em objetivo e nível
func GerarTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dados struct {
			Objetivo string `json:"objetivo"`
			Nivel    string `json:"nivel"`
		}

		if err := json.NewDecoder(r.Body).Decode(&dados); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Consulta otimizada para exercícios
		rows, err := db.Query(`
			SELECT nome FROM exercicios
			WHERE objetivo = $1 AND nivel = $2`,
			dados.Objetivo, dados.Nivel,
		)

		if err != nil {
			http.Error(w, "Erro no banco de dados", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var exercicios []string
		for rows.Next() {
			var nome string
			if err := rows.Scan(&nome); err != nil {
				http.Error(w, "Erro interno", http.StatusInternalServerError)
				return
			}
			exercicios = append(exercicios, nome)
		}

		resposta := Treino{
			Objetivo:   dados.Objetivo,
			Nivel:      dados.Nivel,
			Exercicios: exercicios,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resposta)
	}
}

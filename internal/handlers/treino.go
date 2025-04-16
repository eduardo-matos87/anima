package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// Treino representa uma divisão com exercícios
type Treino struct {
	Divisao    string   `json:"divisao"`
	Exercicios []string `json:"exercicios"`
}

// RespostaTreino é a resposta final da API
type RespostaTreino struct {
	Nivel    string            `json:"nivel"`
	Objetivo string            `json:"objetivo"`
	Dias     int               `json:"dias"`
	Treinos  map[string]Treino `json:"treinos"`
}

// GerarTreino retorna treinos baseados em nível e objetivo
func GerarTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nivel := strings.ToLower(r.URL.Query().Get("nivel"))
		objetivo := strings.ToLower(r.URL.Query().Get("objetivo"))

		treinosPorDivisao := make(map[string]Treino)

		query := `
			SELECT t.divisao, e.nome
			FROM treinos t
			JOIN treino_exercicios te ON t.id = te.treino_id
			JOIN exercicios e ON e.id = te.exercicio_id
			WHERE LOWER(t.nivel) = ? AND LOWER(t.objetivo) = ?
			ORDER BY t.divisao
		`

		rows, err := db.Query(query, nivel, objetivo)
		if err != nil {
			http.Error(w, "Erro ao consultar treinos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var divisao, exercicio string

			if err := rows.Scan(&divisao, &exercicio); err != nil {
				http.Error(w, "Erro ao ler dados do treino", http.StatusInternalServerError)
				return
			}

			if _, exists := treinosPorDivisao[divisao]; !exists {
				treinosPorDivisao[divisao] = Treino{
					Divisao:    divisao,
					Exercicios: []string{},
				}
			}

			treino := treinosPorDivisao[divisao]
			treino.Exercicios = append(treino.Exercicios, exercicio)
			treinosPorDivisao[divisao] = treino
		}

		resposta := RespostaTreino{
			Nivel:    nivel,
			Objetivo: objetivo,
			Dias:     len(treinosPorDivisao),
			Treinos:  treinosPorDivisao,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resposta)
	}
}

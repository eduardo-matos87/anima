package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// Estrutura de resposta para um exercício
type Exercicio struct {
	ID   int64  `json:"id"`
	Nome string `json:"nome"`
}

// Handler que retorna exercícios filtrados por grupo muscular
func ListarExercicios(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		grupo := strings.ToLower(r.URL.Query().Get("grupo"))
		if grupo == "" {
			http.Error(w, "Parâmetro 'grupo' é obrigatório", http.StatusBadRequest)
			return
		}

		// Consulta com JOIN para buscar exercícios pelo nome do grupo
		query := `
			SELECT e.id, e.nome
			FROM exercicios e
			JOIN grupos_musculares gm ON gm.id = e.grupo_id
			WHERE LOWER(gm.nome) = ?
			ORDER BY e.nome
		`

		rows, err := db.Query(query, grupo)
		if err != nil {
			http.Error(w, "Erro ao buscar exercícios", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var exercicios []Exercicio
		for rows.Next() {
			var ex Exercicio
			if err := rows.Scan(&ex.ID, &ex.Nome); err != nil {
				http.Error(w, "Erro ao ler exercício", http.StatusInternalServerError)
				return
			}
			exercicios = append(exercicios, ex)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(exercicios)
	}
}

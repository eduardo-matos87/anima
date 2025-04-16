// Arquivo: internal/handlers/exercicios.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Exercicio model
type Exercicio struct {
	ID      int    `json:"id"`
	Nome    string `json:"nome"`
	GrupoID int    `json:"grupo_id"`
}

// ListarExercicios retorna todos os exercícios (filtra por ?grupo= opcional)
func ListarExercicios(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		grupo := r.URL.Query().Get("grupo")
		var rows *sql.Rows
		var err error
		if grupo != "" {
			rows, err = db.Query(`
				SELECT e.id,e.nome,e.grupo_id 
				FROM exercicios e 
				JOIN grupos_musculares g ON g.id=e.grupo_id 
				WHERE g.nome=?`, grupo)
		} else {
			rows, err = db.Query("SELECT id,nome,grupo_id FROM exercicios")
		}
		if err != nil {
			http.Error(w, "Erro ao buscar exercícios", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []Exercicio
		for rows.Next() {
			var e Exercicio
			rows.Scan(&e.ID, &e.Nome, &e.GrupoID)
			list = append(list, e)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

// Arquivo: internal/handlers/objetivos.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Objetivo model
type Objetivo struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

// ListarObjetivos retorna todos os objetivos
func ListarObjetivos(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id,nome FROM objetivos")
		if err != nil {
			http.Error(w, "Erro ao buscar objetivos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []Objetivo
		for rows.Next() {
			var o Objetivo
			rows.Scan(&o.ID, &o.Nome)
			list = append(list, o)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

// Arquivo: internal/handlers/grupos.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// GrupoMuscular model
type GrupoMuscular struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

// ListarGruposMusculares retorna todos os grupos
func ListarGruposMusculares(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id,nome FROM grupos_musculares")
		if err != nil {
			http.Error(w, "Erro ao buscar grupos musculares", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []GrupoMuscular
		for rows.Next() {
			var g GrupoMuscular
			rows.Scan(&g.ID, &g.Nome)
			list = append(list, g)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

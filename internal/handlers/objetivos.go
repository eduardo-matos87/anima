package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// GrupoMuscular representa um grupo muscular cadastrado.
type GrupoMuscular struct {
	ID   int64  `json:"id"`
	Nome string `json:"nome"`
}

// ListarGruposMusculares consulta e retorna os grupos musculares cadastrados no banco.
func ListarGruposMusculares(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, nome FROM grupos_musculares ORDER BY id")
		if err != nil {
			http.Error(w, "Erro ao buscar grupos musculares", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var grupos []GrupoMuscular
		for rows.Next() {
			var g GrupoMuscular
			if err := rows.Scan(&g.ID, &g.Nome); err != nil {
				http.Error(w, "Erro ao ler grupos musculares", http.StatusInternalServerError)
				return
			}
			grupos = append(grupos, g)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(grupos)
	}
}

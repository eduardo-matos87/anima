package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
)

// Objetivo representa um objetivo de treino cadastrado no banco.
type Objetivo struct {
    ID   int64  `json:"id"`
    Nome string `json:"nome"`
}

// ListarObjetivos consulta e retorna os objetivos cadastrados.
func ListarObjetivos(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.Query("SELECT id, nome FROM objetivos ORDER BY id")
        if err != nil {
            http.Error(w, "Erro ao buscar objetivos", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var objetivos []Objetivo
        for rows.Next() {
            var o Objetivo
            if err := rows.Scan(&o.ID, &o.Nome); err != nil {
                http.Error(w, "Erro ao ler objetivos", http.StatusInternalServerError)
                return
            }
            objetivos = append(objetivos, o)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(objetivos)
    }
}

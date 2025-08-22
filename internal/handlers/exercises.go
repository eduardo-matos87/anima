package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type ExerciseItem struct {
	ID    int64  `json:"id"`
	Nome  string `json:"nome"`
	Grupo string `json:"grupo"`
}

type ListExercisesResp struct {
	Items []ExerciseItem `json:"items"`
}

func ListExercises(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		grupo := strings.TrimSpace(r.URL.Query().Get("grupo"))

		base := `SELECT id, nome, grupo FROM exercises WHERE 1=1`
		args := []any{}
		if q != "" {
			base += ` AND unaccent(lower(nome)) LIKE unaccent(lower($1))`
			args = append(args, "%"+q+"%")
		}
		if grupo != "" {
			base += ` AND lower(grupo) = lower($` + string('1'+len(args)) + `)`
			args = append(args, grupo)
		}
		base += ` ORDER BY nome LIMIT 100`

		rows, err := db.Query(base, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var items []ExerciseItem
		for rows.Next() {
			var it ExerciseItem
			if err := rows.Scan(&it.ID, &it.Nome, &it.Grupo); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, it)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListExercisesResp{Items: items})
	}
}

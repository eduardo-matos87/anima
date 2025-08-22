package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
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

func ListExercises(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		grupo := strings.TrimSpace(r.URL.Query().Get("grupo"))

		// ⚠️ Lê da tabela EN: exercises
		base := `SELECT id, name AS nome, muscle_group AS grupo FROM exercises WHERE 1=1`
		args := []any{}

		if q != "" {
			base += ` AND name ILIKE ` + place(len(args)+1)
			args = append(args, "%"+q+"%")
		}
		if grupo != "" {
			base += ` AND lower(muscle_group) = lower(` + place(len(args)+1) + `)`
			args = append(args, grupo)
		}
		base += ` ORDER BY name LIMIT 100`

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
		_ = json.NewEncoder(w).Encode(ListExercisesResp{Items: items})
	})
}

func place(i int) string { return "$" + strconv.Itoa(i) }

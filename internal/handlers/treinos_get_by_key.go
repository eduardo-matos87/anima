package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func GetTreinoByKey(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		key := strings.TrimPrefix(r.URL.Path, "/api/treinos/by-key/")
		if key == "" || strings.Contains(key, "/") {
			http.NotFound(w, r)
			return
		}

		var d TreinoDetail
		var coach sql.NullString
		err := db.QueryRow(`
			SELECT id, objetivo, nivel, dias, divisao, coach_notes
			FROM treinos
			WHERE treino_key = $1
			LIMIT 1
		`, key).Scan(&d.ID, &d.Objetivo, &d.Nivel, &d.Dias, &d.Divisao, &coach)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if coach.Valid {
			s := coach.String
			d.CoachNotes = &s
		}

		rows, err := db.Query(`
			SELECT te.exercicio_id, e.name AS nome, e.muscle_group AS grupo, te.series, te.repeticoes
			FROM treino_exercicios te
			JOIN exercises e ON e.id = te.exercicio_id
			WHERE te.treino_id = $1
			ORDER BY te.id
		`, d.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var it TreinoItem
			if err := rows.Scan(&it.ExercicioID, &it.Nome, &it.Grupo, &it.Series, &it.Repeticoes); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			d.Exercicios = append(d.Exercicios, it)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(d)
	})
}

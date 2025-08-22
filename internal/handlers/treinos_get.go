package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type TreinoItem struct {
	ExercicioID int64  `json:"exercicio_id"`
	Nome        string `json:"nome"`
	Grupo       string `json:"grupo"`
	Series      int    `json:"series"`
	Repeticoes  string `json:"repeticoes"`
}

type TreinoDetail struct {
	ID         int64        `json:"id"`
	Objetivo   string       `json:"objetivo"`
	Nivel      string       `json:"nivel"`
	Dias       int          `json:"dias"`
	Divisao    string       `json:"divisao"`
	CoachNotes *string      `json:"coach_notes,omitempty"`
	Exercicios []TreinoItem `json:"exercicios"`
}

func GetTreinoByID(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/treinos/")
		if idStr == "" || strings.Contains(idStr, "/") {
			http.NotFound(w, r)
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "id inválido", http.StatusBadRequest)
			return
		}

		var d TreinoDetail
		d.ID = id
		var coach sql.NullString
		err = db.QueryRow(`SELECT objetivo, nivel, dias, divisao, coach_notes FROM treinos WHERE id=$1`, id).
			Scan(&d.Objetivo, &d.Nivel, &d.Dias, &d.Divisao, &coach)
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
		`, id)
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

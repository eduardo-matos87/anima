package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type SaveTreinoReq struct {
	TreinoID   string `json:"treino_id"`
	Exercicios []struct {
		ExerciseID int64  `json:"exercise_id"`
		Series     int    `json:"series"`
		Repeticoes string `json:"repeticoes"`
	} `json:"exercicios"`
}

func SaveTreino(_ *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req SaveTreinoReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json inválido", http.StatusBadRequest)
			return
		}
		// TODO: INSERT em treinos e treino_exercicios
		w.WriteHeader(http.StatusCreated)
	})
}

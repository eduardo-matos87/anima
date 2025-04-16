package handlers

import (
	"encoding/json"
	"net/http"
)

type Treino struct {
	Dia       string   `json:"dia"`
	Exercicios []string `json:"exercicios"`
}

func GerarTreino(w http.ResponseWriter, r *http.Request) {
	treino := Treino{
		Dia:       "Segunda",
		Exercicios: []string{"Supino reto", "Supino inclinado", "Crucifixo", "Tríceps testa"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(treino)
}
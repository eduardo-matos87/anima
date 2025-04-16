package handlers

import (
	"encoding/json"
	"net/http"
)

type Treino struct {
	Nivel     string   `json:"nivel"`
	Objetivo  string   `json:"objetivo"`
	Exercicios []string `json:"exercicios"`
}

func GerarTreino(w http.ResponseWriter, r *http.Request) {
	nivel := r.URL.Query().Get("nivel")
	objetivo := r.URL.Query().Get("objetivo")

	exercicios := []string{"Flexão", "Agachamento", "Prancha"}

	if nivel == "intermediario" && objetivo == "hipertrofia" {
		exercicios = []string{"Supino reto", "Puxada alta", "Agachamento livre", "Remada"}
	}

	treino := Treino{
		Nivel:     nivel,
		Objetivo:  objetivo,
		Exercicios: exercicios,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(treino)
}
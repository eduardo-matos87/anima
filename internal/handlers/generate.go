package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type GenerateTreinoReq struct {
	Objetivo string `json:"objetivo"`
	Nivel    string `json:"nivel"`
	Divisao  string `json:"divisao"`
}

type ExercicioResp struct {
	Nome        string `json:"nome"`
	Series      int    `json:"series"`
	Repeticoes  string `json:"repeticoes"`
	DescansoSeg int    `json:"descanso_seg"`
}

type GenerateTreinoResp struct {
	TreinoID   string          `json:"treino_id"`
	Exercicios []ExercicioResp `json:"exercicios"`
}

// GenerateTreino retorna um http.Handler (compatível com main.go)
func GenerateTreino(_ *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req GenerateTreinoReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json inválido", http.StatusBadRequest)
			return
		}

		// Mock simples — depois trocamos por SELECTs em exercises
		pool := []string{
			"Agachamento Livre", "Supino Reto", "Levantamento Terra",
			"Remada Curvada", "Desenvolvimento Militar", "Puxada na Frente",
			"Rosca Direta", "Tríceps Testa", "Elevação Lateral",
		}

		exs := make([]ExercicioResp, 0, 6)
		for i := 0; i < 6 && i < len(pool); i++ {
			exs = append(exs, ExercicioResp{
				Nome:        pool[i],
				Series:      3 + (i % 2), // 3 ou 4
				Repeticoes:  "8-12",
				DescansoSeg: 60 + (i%3)*30, // 60/90/120
			})
		}

		resp := GenerateTreinoResp{
			TreinoID:   time.Now().Format("20060102T150405"),
			Exercicios: exs,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

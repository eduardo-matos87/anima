package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

type SaveTreinoReq struct {
	TreinoID   string `json:"treino_id"` // opcional: id lógico gerado no /generate
	Objetivo   string `json:"objetivo"`  // bate com treinos.objetivo
	Nivel      string `json:"nivel"`     // bate com treinos.nivel
	Dias       int    `json:"dias"`      // ⚠️ obrigatório (treinos.dias é NOT NULL)
	Divisao    string `json:"divisao"`   // bate com treinos.divisao
	Exercicios []struct {
		ExercicioID int64  `json:"exercicio_id"` // ⚠️ casa com treino_exercicios.exercicio_id (FK -> exercises.id)
		Series      int    `json:"series"`       // precisa da migration 011
		Repeticoes  string `json:"repeticoes"`   // precisa da migration 011
	} `json:"exercicios"`
}

type SaveTreinoResp struct {
	ID       int64  `json:"id"`
	TreinoID string `json:"treino_id"`
}

func SaveTreino(db *sql.DB) http.Handler {
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
		if req.Dias <= 0 || req.Nivel == "" || req.Objetivo == "" || req.Divisao == "" {
			http.Error(w, "campos obrigatórios: objetivo, nivel, dias, divisao", http.StatusBadRequest)
			return
		}
		if len(req.Exercicios) == 0 {
			http.Error(w, "exercicios obrigatórios", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()

		// INSERT em treinos
		var treinoDBID int64
		err = tx.QueryRow(`
			INSERT INTO treinos (nivel, objetivo, dias, divisao, treino_key)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''))
			RETURNING id
		`, req.Nivel, req.Objetivo, req.Dias, req.Divisao, req.TreinoID).Scan(&treinoDBID)

		if err != nil {
			http.Error(w, "erro ao criar treino: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// INSERT itens
		stmt, err := tx.Prepare(`
			INSERT INTO treino_exercicios (treino_id, exercicio_id, series, repeticoes)
			VALUES ($1, $2, $3, $4)
		`)
		if err != nil {
			http.Error(w, "erro preparando insert de exercícios: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		for _, e := range req.Exercicios {
			if e.ExercicioID == 0 || e.Series <= 0 || e.Repeticoes == "" {
				err = errors.New("exercicio_id/series/repeticoes inválidos")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if _, err = stmt.Exec(treinoDBID, e.ExercicioID, e.Series, e.Repeticoes); err != nil {
				http.Error(w, "erro ao inserir exercício: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err = tx.Commit(); err != nil {
			http.Error(w, "erro ao salvar treino: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(SaveTreinoResp{
			ID:       treinoDBID,
			TreinoID: req.TreinoID,
		})
	})
}

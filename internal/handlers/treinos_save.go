package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
)

type SaveTreinoReq struct {
	TreinoID   string              `json:"treino_id"` // key lógica opcional/única
	Objetivo   string              `json:"objetivo"`
	Nivel      string              `json:"nivel"`
	Dias       int                 `json:"dias"`
	Divisao    string              `json:"divisao"`
	CoachNotes *string             `json:"coach_notes,omitempty"`
	Exercicios []SaveTreinoItemReq `json:"exercicios"`
}

type SaveTreinoItemReq struct {
	ExercicioID int64  `json:"exercicio_id"`
	Series      int    `json:"series"`
	Repeticoes  string `json:"repeticoes"`
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
		if req.Objetivo == "" || req.Nivel == "" || req.Dias <= 0 || req.Divisao == "" || len(req.Exercicios) == 0 {
			http.Error(w, "campos obrigatórios ausentes", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = tx.Rollback() }()

		var treinoDBID int64
		err = tx.QueryRow(`
			INSERT INTO treinos (nivel, objetivo, dias, divisao, treino_key, coach_notes)
			VALUES ($1, $2, $3, $4, NULLIF($5,''), NULLIF($6,''))
			RETURNING id
		`, req.Nivel, req.Objetivo, req.Dias, req.Divisao, req.TreinoID, optStr(req.CoachNotes)).Scan(&treinoDBID)
		if err != nil {
			// se for violação de unicidade (treino_key único), responder 409
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
				http.Error(w, "treino_id já existe", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stmt, err := tx.Prepare(`
			INSERT INTO treino_exercicios (treino_id, exercicio_id, series, repeticoes)
			VALUES ($1, $2, $3, $4)
		`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		for _, it := range req.Exercicios {
			if it.ExercicioID <= 0 || it.Series <= 0 || it.Repeticoes == "" {
				http.Error(w, "item inválido em exercicios", http.StatusBadRequest)
				return
			}
			if _, err := stmt.Exec(treinoDBID, it.ExercicioID, it.Series, it.Repeticoes); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(SaveTreinoResp{ID: treinoDBID, TreinoID: req.TreinoID})
	})
}

// retorna nil se ponteiro vazio/nil – pra evitar gravar ""
func optStr(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

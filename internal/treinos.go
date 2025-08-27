package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
)

type GenerateRequest struct {
	Goal  string `json:"goal"`  // hypertrophy|strength|fatloss (por enquanto informativo)
	Level string `json:"level"` // beginner|intermediate|advanced (não usado diretamente ainda)
	Days  int    `json:"days"`  // 3,4,5...
	Split string `json:"split"` // push-pull-legs|upper-lower|full-body
}

type Exercise struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MuscleGroup string `json:"muscle_group"`
}

type DayPlan struct {
	Day   int        `json:"day"`
	Focus string     `json:"focus"`
	Items []Exercise `json:"items"`
}

type GenerateResponse struct {
	Program []DayPlan `json:"program"`
}

func GenerateTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.Level == "" {
			req.Level = "beginner"
		}
		if req.Days <= 0 {
			req.Days = 3
		}
		if req.Split == "" {
			req.Split = "push-pull-legs"
		}

		foci := buildSplit(req.Split, req.Days)
		if len(foci) == 0 {
			http.Error(w, "invalid split/days", http.StatusBadRequest)
			return
		}

		resp := GenerateResponse{Program: make([]DayPlan, 0, len(foci))}
		for i, focus := range foci {
			groupNames := groupsForFocusPT(focus) // nomes conforme tabela `grupos`
			items, err := pickFromDB(ctx, db, groupNames)
			if err != nil {
				http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			resp.Program = append(resp.Program, DayPlan{
				Day:   i + 1,
				Focus: focus,
				Items: items,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func buildSplit(split string, days int) []string {
	switch strings.ToLower(split) {
	case "push-pull-legs":
		seq := []string{"push", "pull", "legs"}
		out := make([]string, 0, days)
		for i := 0; i < days; i++ {
			out = append(out, seq[i%3])
		}
		return out
	case "upper-lower":
		seq := []string{"upper", "lower"}
		out := make([]string, 0, days)
		for i := 0; i < days; i++ {
			out = append(out, seq[i%2])
		}
		return out
	case "full-body":
		out := make([]string, days)
		for i := range out {
			out[i] = "full"
		}
		return out
	default:
		return nil
	}
}

// mapeia focos -> nomes na sua tabela `grupos` (PT-BR)
func groupsForFocusPT(focus string) []string {
	switch focus {
	case "push":
		return []string{"Peito", "Ombros", "Tríceps"}
	case "pull":
		return []string{"Costas", "Bíceps"}
	case "legs":
		return []string{"Pernas", "Abdômen"}
	case "upper":
		return []string{"Peito", "Costas", "Ombros", "Bíceps", "Tríceps"}
	case "lower":
		return []string{"Pernas", "Abdômen"}
	case "full":
		return []string{"Peito", "Costas", "Pernas", "Ombros", "Bíceps", "Tríceps", "Abdômen", "Cardio"}
	default:
		return []string{}
	}
}

// escolhe exercícios por grupo, 1-2 por grupo, ordem aleatória
func pickFromDB(ctx context.Context, db *sql.DB, groupNames []string) ([]Exercise, error) {
	if len(groupNames) == 0 {
		return nil, errors.New("no groups")
	}
	// Seleciona até 2 exercícios por grupo, misturados; limite total 7
	// (ajuste conforme preferir)
	const q = `
WITH gs AS (
  SELECT id, nome FROM grupos WHERE nome = ANY($1)
),
ex AS (
  SELECT e.id, e.nome, g.nome AS grupo,
         ROW_NUMBER() OVER (PARTITION BY g.id ORDER BY random()) AS rn
  FROM exercicios e
  JOIN gs g ON g.id = e.grupo_id
)
SELECT id, nome, grupo
FROM ex
WHERE rn <= 2
ORDER BY random()
LIMIT 7;
`
	rows, err := db.QueryContext(ctx, q, pq.Array(groupNames))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Exercise
	for rows.Next() {
		var e Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.MuscleGroup); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

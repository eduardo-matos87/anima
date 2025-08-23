package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// ====== Tipos de request/response

type GenerateReq struct {
	Objetivo string `json:"objetivo"`            // ex: "hipertrofia", "emagrecimento", "forca", "resistencia"
	Nivel    string `json:"nivel"`               // ex: "iniciante", "intermediario", "avancado"
	Divisao  string `json:"divisao"`             // ex: "fullbody" (por ora não influencia a lógica)
	Dias     int    `json:"dias,omitempty"`      // default 3
	Persist  *bool  `json:"persist,omitempty"`   // default: true (persiste automaticamente)
	TreinoID string `json:"treino_id,omitempty"` // opcional para fixar a chave lógica (única)
}

type GeneratedExercise struct {
	ExercicioID int    `json:"exercicio_id"`
	Nome        string `json:"nome"`
	Series      int    `json:"series"`
	Repeticoes  string `json:"repeticoes"`
}

type GenerateResp struct {
	ID         *int                `json:"id,omitempty"` // presente se persistido
	TreinoID   string              `json:"treino_id"`    // key lógica (gerada se não informada)
	Exercicios []GeneratedExercise `json:"exercicios"`   // plano gerado
	CoachNotes string              `json:"coach_notes,omitempty"`
}

// ====== Handler

func GenerateTreino(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req GenerateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json inválido", http.StatusBadRequest)
			return
		}
		// sane defaults
		req.Objetivo = strings.TrimSpace(strings.ToLower(req.Objetivo))
		req.Nivel = strings.TrimSpace(strings.ToLower(req.Nivel))
		req.Divisao = strings.TrimSpace(strings.ToLower(req.Divisao))
		if req.Dias <= 0 {
			req.Dias = 3
		}
		if req.Objetivo == "" || req.Nivel == "" || req.Divisao == "" {
			http.Error(w, "campos obrigatórios: objetivo, nivel, divisao", http.StatusBadRequest)
			return
		}
		persist := true
		if req.Persist != nil {
			persist = *req.Persist
		}

		uid := getUserID(r) // permanece igual ao restante do projeto

		// ===== Perfil + métricas =====
		prof, _ := loadUserProfile(r.Context(), db, uid)
		if wkg, _ := latestWeight(r.Context(), db, uid); wkg != nil {
			prof.WeightKG = wkg
		}

		// ===== Seleção de exercícios a partir da tabela exercises =====
		exs, err := buildPlanFromCatalog(r.Context(), db, req)
		if err != nil {
			http.Error(w, "falha ao montar plano: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(exs) == 0 {
			http.Error(w, "não há exercícios cadastrados no catálogo", http.StatusConflict)
			return
		}

		// ===== Coach notes (somente se use_ai=true no perfil) =====
		coach := ""
		if prof.UseAI == nil || *prof.UseAI {
			coach = buildCoachNotes(req, prof)
		}

		// ===== Persistência opcional =====
		key := req.TreinoID
		if key == "" {
			key = "gen-" + time.Now().Format("20060102T150405")
		}

		var insertedID *int
		if persist {
			id, err := persistPlan(r.Context(), db, key, req, coach, exs)
			if err != nil {
				// conflito de treino_id único, tente outro automático
				if strings.Contains(err.Error(), "duplicate key") {
					key = "gen-" + time.Now().Format("20060102T150405.000")
					id, err = persistPlan(r.Context(), db, key, req, coach, exs)
				}
			}
			if err != nil {
				http.Error(w, "falha ao salvar treino: "+err.Error(), http.StatusInternalServerError)
				return
			}
			insertedID = &id
			w.WriteHeader(http.StatusCreated)
		} else {
			// preview
			w.WriteHeader(http.StatusOK)
		}

		resp := GenerateResp{
			ID:         insertedID,
			TreinoID:   key,
			Exercicios: exs,
			CoachNotes: coach,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// ====== Lógica de plano (consulta em exercises)

func buildPlanFromCatalog(ctx context.Context, db *sql.DB, req GenerateReq) ([]GeneratedExercise, error) {
	// alvo: 6 exercícios no plano base
	const target = 6

	// 1) Preferência por nível
	es, err := queryExercises(ctx, db, req.Nivel, target)
	if err != nil {
		return nil, err
	}
	// 2) Se faltou, complementa sem filtro de nível
	if len(es) < target {
		rest, err := queryExercises(ctx, db, "", target-len(es))
		if err != nil {
			return nil, err
		}
		es = append(es, rest...)
	}
	if len(es) == 0 {
		return nil, nil
	}
	if len(es) > target {
		es = es[:target]
	}

	// 3) Atribui séries/reps conforme objetivo
	var reps string
	switch req.Objetivo {
	case "hipertrofia":
		reps = "8-12"
	case "emagrecimento":
		reps = "12-15"
	case "forca", "força":
		reps = "4-6"
	case "resistencia", "resistência":
		reps = "12-20"
	default:
		reps = "8-12"
	}

	out := make([]GeneratedExercise, 0, len(es))
	for i, it := range es {
		series := 3
		// pequeno ajuste por “intensidade”
		if req.Objetivo == "hipertrofia" && i%2 == 1 {
			series = 4
		}
		out = append(out, GeneratedExercise{
			ExercicioID: it.id,
			Nome:        it.name,
			Series:      series,
			Repeticoes:  reps,
		})
	}
	return out, nil
}

type exRow struct {
	id   int
	name string
}

func queryExercises(ctx context.Context, db *sql.DB, nivel string, limit int) ([]exRow, error) {
	args := []any{}
	q := `
		SELECT id, name
		FROM exercises
	`
	if nivel != "" {
		q += ` WHERE lower(difficulty) = $1 `
		args = append(args, strings.ToLower(nivel))
	}
	q += ` ORDER BY id ASC LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []exRow
	for rows.Next() {
		var r exRow
		if err := rows.Scan(&r.id, &r.name); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ====== Persistência

func persistPlan(ctx context.Context, db *sql.DB, key string, req GenerateReq, coach string, plan []GeneratedExercise) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var treinoID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO treinos (objetivo, nivel, dias, divisao, treino_key, coach_notes)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`, req.Objetivo, req.Nivel, req.Dias, req.Divisao, key, nullIfEmpty(coach)).Scan(&treinoID)
	if err != nil {
		return 0, err
	}

	for _, ex := range plan {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO treino_exercicios (treino_id, exercicio_id, series, repeticoes)
			VALUES ($1,$2,$3,$4)
		`, treinoID, ex.ExercicioID, ex.Series, ex.Repeticoes)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return treinoID, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// ====== Coach notes (reaproveita helpers do projeto)

func buildCoachNotes(req GenerateReq, p userProfile) string {
	obj := req.Objetivo
	if obj == "" && p.TrainingGoal != nil && *p.TrainingGoal != "" {
		obj = *p.TrainingGoal
	}

	// idade
	var idade *int
	now := time.Now()
	if p.BirthDate != nil {
		y := computeAge(*p.BirthDate, now)
		idade = &y
	}

	// IMC
	var imc *float64
	if p.HeightCM != nil && *p.HeightCM > 0 && p.WeightKG != nil && *p.WeightKG > 0 {
		hm := float64(*p.HeightCM) / 100.0
		v := *p.WeightKG / (hm * hm)
		v = math.Round(v*100) / 100
		imc = &v
	}

	txt := fmt.Sprintf("Plano %s (%s), divisão %s. ",
		valOr("geral", obj), valOr("nível indefinido", req.Nivel), valOr("indefinida", req.Divisao))

	if p.HeightCM != nil {
		txt += fmt.Sprintf("Altura %dcm. ", *p.HeightCM)
	}
	if p.WeightKG != nil {
		txt += fmt.Sprintf("Peso %.1fkg. ", *p.WeightKG)
	}
	if idade != nil {
		txt += fmt.Sprintf("Idade %d. ", *idade)
	}
	if imc != nil {
		txt += fmt.Sprintf("IMC=%.2f. ", *imc)
	}

	switch obj {
	case "hipertrofia":
		txt += "Foque em progressão de carga com técnica sólida; 8–12 reps nos compostos; sono ≥ 7h."
	case "emagrecimento":
		txt += "Aumente densidade do treino (descanso curto) e mantenha leve déficit calórico."
	case "forca", "força":
		txt += "Priorize compostos pesados; séries curtas e descanso maior; monitore a técnica."
	case "resistencia", "resistência":
		txt += "Volume moderado/alto, cadência controlada e constância semanal."
	default:
		txt += "Mantenha técnica perfeita, aquecimento e progressão gradual."
	}
	if idade != nil && *idade >= 40 {
		txt += " Aqueça bem ombros/quadril; evite picos de carga abruptos."
	}
	return txt
}

func valOr(def, v string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

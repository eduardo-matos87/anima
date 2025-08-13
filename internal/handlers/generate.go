package handlers

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ======== Tipos de request/response ========

type GenerateRequest struct {
	Goal         string   `json:"goal"`          // hipertrofia|forca|resistencia
	Level        string   `json:"level"`         // beginner|intermediate|advanced
	DaysPerWeek  int      `json:"days_per_week"` // 2..6
	Equipment    []string `json:"equipment"`     // ["halter","barra","maquina","livre"]
	Restrictions []string `json:"restrictions"`  // ex: ["ombro","joelho"]
}

type WorkoutExercise struct {
	DayIndex    int    `json:"day_index"`
	ExerciseID  int    `json:"exercise_id"`
	Name        string `json:"name"`
	Sets        int    `json:"sets"`
	Reps        string `json:"reps"`
	RestSeconds int    `json:"rest_seconds"`
	Tempo       string `json:"tempo,omitempty"`
}

type GenerateResponse struct {
	Plan struct {
		Goal        string            `json:"goal"`
		Level       string            `json:"level"`
		DaysPerWeek int               `json:"days_per_week"`
		Split       []string          `json:"split"`
		Items       []WorkoutExercise `json:"items"`
		Notes       string            `json:"notes"`
	} `json:"plan"`
}

type Exercise struct {
	ID            int
	Name          string
	PrimaryMuscle string
	Difficulty    string
	Equipment     string
	Unilateral    bool
}

// ======== Handler ========

type GenerateHandler struct {
	DB *pgxpool.Pool
}

func NewGenerateHandler(db *pgxpool.Pool) *GenerateHandler {
	return &GenerateHandler{DB: db}
}

func (h *GenerateHandler) GenerateWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	normalizeReq(&req)

	rand.Seed(time.Now().UnixNano())

	// carrega exercícios do banco
	pool := h.fetchExercises(r.Context())

	sets, reps, rest := scheme(req.Goal, req.Level)
	split := chooseSplit(req.DaysPerWeek)

	items := make([]WorkoutExercise, 0, 64)
	for dayIdx, label := range split {
		muscles := muscleBucket(label)
		candidates := filterExercises(pool, muscles, req.Equipment, req.Level, req.Restrictions)

		// regra simples: 5 exercícios por dia; se tiver "core" explícito, usa 4
		pick := 5
		if strings.Contains(strings.ToLower(label), "core") {
			pick = 4
		}
		chosen := pickN(candidates, pick)

		for _, ex := range chosen {
			items = append(items, WorkoutExercise{
				DayIndex:    dayIdx + 1,
				ExerciseID:  ex.ID,
				Name:        ex.Name,
				Sets:        sets,
				Reps:        reps,
				RestSeconds: rest,
			})
		}
	}

	var resp GenerateResponse
	resp.Plan.Goal = req.Goal
	resp.Plan.Level = req.Level
	resp.Plan.DaysPerWeek = req.DaysPerWeek
	resp.Plan.Split = split
	resp.Plan.Items = items
	resp.Plan.Notes = refineNotesWithAI(req.Goal, req.Level, split)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
	log.Printf("Generated plan: goal=%s level=%s days=%d items=%d",
		req.Goal, req.Level, req.DaysPerWeek, len(items))
}

// ======== DB ========

func (h *GenerateHandler) fetchExercises(ctx context.Context) []Exercise {
	rows, err := h.DB.Query(ctx, `
		SELECT id, name, primary_muscle, difficulty, equipment, is_unilateral
		FROM exercises
	`)
	if err != nil {
		return []Exercise{}
	}
	defer rows.Close()

	out := make([]Exercise, 0, 128)
	for rows.Next() {
		var e Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.PrimaryMuscle, &e.Difficulty, &e.Equipment, &e.Unilateral); err == nil {
			out = append(out, e)
		}
	}
	return out
}

// ======== Regras ========

func normalizeReq(req *GenerateRequest) {
	if req.DaysPerWeek < 2 {
		req.DaysPerWeek = 3
	}
	if req.Level == "" {
		req.Level = "beginner"
	}
	if req.Goal == "" {
		req.Goal = "hipertrofia"
	}
	if len(req.Equipment) == 0 {
		req.Equipment = []string{"halter", "barra", "maquina", "livre"}
	}
}

func scheme(goal, level string) (sets int, reps string, rest int) {
	switch strings.ToLower(goal) {
	case "forca":
		return 5, "3-5", 150
	case "resistencia":
		return 3, "15-20", 45
	default: // hipertrofia
		if strings.ToLower(level) == "beginner" {
			return 3, "10-12", 75
		}
		return 4, "8-12", 90
	}
}

func chooseSplit(days int) []string {
	switch days {
	case 2:
		return []string{"Full Body A", "Full Body B"}
	case 3:
		return []string{"Peito+Tríceps", "Costas+Bíceps", "Pernas+Ombros+Core"}
	case 4:
		return []string{"Superior A", "Inferior A", "Superior B", "Inferior B"}
	case 5:
		return []string{"Peito", "Costas", "Pernas", "Ombros", "Braços+Core"}
	default:
		return []string{"Peito+Tríceps", "Costas+Bíceps", "Pernas", "Ombros", "Braços", "Core"}
	}
}

func muscleBucket(label string) []string {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "peito"):
		return []string{"peito"}
	case strings.Contains(l, "costas"):
		return []string{"costas"}
	case strings.Contains(l, "pernas"):
		return []string{"pernas", "panturrilha"}
	case strings.Contains(l, "ombros"):
		return []string{"ombros"}
	case strings.Contains(l, "braços"):
		return []string{"biceps", "triceps"}
	case strings.Contains(l, "core"):
		return []string{"core"}
	default:
		return []string{"peito", "costas", "pernas", "ombros", "biceps", "triceps", "core", "panturrilha"}
	}
}

func contains(xs []string, s string) bool {
	s = strings.ToLower(s)
	for _, x := range xs {
		if strings.ToLower(x) == s {
			return true
		}
	}
	return false
}

func filterExercises(pool []Exercise, muscles []string, eqAllowed []string, level string, restrictions []string) []Exercise {
	ok := make([]Exercise, 0)
	for _, ex := range pool {
		if !contains(eqAllowed, ex.Equipment) && ex.Equipment != "livre" {
			continue
		}
		bad := false
		for _, r := range restrictions {
			if strings.Contains(strings.ToLower(ex.PrimaryMuscle), strings.ToLower(r)) {
				bad = true
				break
			}
		}
		if bad {
			continue
		}
		if ex.Difficulty == "advanced" && strings.ToLower(level) == "beginner" {
			continue
		}
		for _, m := range muscles {
			if strings.ToLower(ex.PrimaryMuscle) == strings.ToLower(m) {
				ok = append(ok, ex)
				break
			}
		}
	}
	return ok
}

func pickN(xs []Exercise, n int) []Exercise {
	if len(xs) <= n {
		return xs
	}
	rand.Shuffle(len(xs), func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
	return xs[:n]
}

func refineNotesWithAI(goal, level string, split []string) string {
	return "Aqueça 5–8 min; técnica primeiro; quando atingir o topo de reps, aumente carga em ~2,5 kg; durma 7–8h."
}

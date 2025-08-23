package services

import (
	"errors"
	"fmt"
	"strings"
)

type Division string

const (
	DivFullBody   Division = "fullbody"
	DivUpperLower Division = "upper-lower"
	DivPPL        Division = "ppl"
)

type GenerateInput struct {
	Persist          bool            `json:"persist"`
	UseAI            bool            `json:"use_ai"`
	Division         Division        `json:"division"`
	DaysPerWeek      int             `json:"days_per_week"`
	TargetGoal       string          `json:"target_goal"`
	IncludeMuscles   []string        `json:"include_muscles"`
	ExcludeExercises []string        `json:"exclude_exercises"`
	Equipment        map[string]bool `json:"equipment"`
	UserID           string          `json:"-"`
}

type Exercise struct {
	ID          string // slug/uuid/etc
	Name        string
	MuscleGroup string   // "chest","back","legs","shoulders","arms","glutes","core","biceps","triceps","core", etc
	Pattern     string   // "compound","isolation","push","pull","squat","hinge", etc
	Equipment   []string // ["barbell","dumbbells","machines",...]
}

// TreinoModel descreve o plano gerado em memória (sem persistir ainda)
// É o “contrato” para os handlers salvarem depois.
type TreinoModel struct {
	Notes    string
	Sessions []SessionModel
}

type SessionModel struct {
	Label       string // "Upper Day 1", "Pull Day 2", ...
	DivisionDay string // "full","upper","lower","push","pull","legs"
	DayIndex    int
	Items       []SessionItem
}

type SessionItem struct {
	ExerciseID string
	OrderIndex int
	Series     int
	Reps       string
}

// Generator é puro (não depende de DB). Quem chama injeta a lista de exercícios disponível.
type Generator struct{}

func NewGenerator() *Generator { return &Generator{} }

func (g *Generator) Generate(in GenerateInput, available []Exercise) (*TreinoModel, error) {
	if in.DaysPerWeek < 1 || in.DaysPerWeek > 7 {
		return nil, errors.New("days_per_week must be between 1 and 7")
	}
	if in.Division == "" {
		in.Division = DivFullBody
	}
	// Filtra por include/exclude/equipment
	filt := filterExercises(available, in.IncludeMuscles, in.ExcludeExercises, in.Equipment)
	if len(filt) == 0 {
		return nil, errors.New("no exercises available after filters")
	}

	plan := buildPlan(in.Division, in.DaysPerWeek)
	out := &TreinoModel{
		Notes: fmt.Sprintf("Plano %s %dd/sem · goal=%s", in.Division, in.DaysPerWeek, in.TargetGoal),
	}

	for i, div := range plan {
		label := fmt.Sprintf("%s Day %d", strings.Title(div), i+1)
		items := pickSessionSet(filt, div, in.TargetGoal)
		// Indexação/ordem
		for idx := range items {
			items[idx].OrderIndex = idx + 1
		}
		out.Sessions = append(out.Sessions, SessionModel{
			Label:       label,
			DivisionDay: div,
			DayIndex:    i + 1,
			Items:       items,
		})
	}
	return out, nil
}

func buildPlan(div Division, d int) []string {
	s := make([]string, d)
	switch div {
	case DivFullBody:
		for i := 0; i < d; i++ {
			s[i] = "full"
		}
	case DivUpperLower:
		for i := 0; i < d; i++ {
			if i%2 == 0 {
				s[i] = "upper"
			} else {
				s[i] = "lower"
			}
		}
	case DivPPL:
		cycle := []string{"push", "pull", "legs"}
		for i := 0; i < d; i++ {
			s[i] = cycle[i%3]
		}
	default:
		for i := 0; i < d; i++ {
			s[i] = "full"
		}
	}
	return s
}

func filterExercises(in []Exercise, includeMuscles, excludeSlugs []string, equipment map[string]bool) []Exercise {
	include := make(map[string]bool)
	for _, m := range includeMuscles {
		include[strings.ToLower(m)] = true
	}
	exclude := make(map[string]bool)
	for _, s := range excludeSlugs {
		exclude[strings.ToLower(s)] = true
	}

	var out []Exercise
	for _, e := range in {
		if e.ID == "" {
			continue
		}
		if exclude[strings.ToLower(e.ID)] {
			continue
		}
		if len(include) > 0 && !include[strings.ToLower(e.MuscleGroup)] {
			continue
		}
		if !equipmentOK(e, equipment) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func equipmentOK(e Exercise, eq map[string]bool) bool {
	if len(eq) == 0 {
		return true // nenhum filtro
	}
	if len(e.Equipment) == 0 {
		return true // assume livre
	}
	for _, need := range e.Equipment {
		if eq[strings.ToLower(need)] {
			return true
		}
	}
	return false
}

// Heurística simples: 1 composto principal + 2 acessórios coerentes com a divisão.
func pickSessionSet(pool []Exercise, div string, goal string) []SessionItem {
	var chosen []Exercise
	switch div {
	case "upper":
		addCompound(&chosen, pool, "chest")
		addCompound(&chosen, pool, "back")
		addAny(&chosen, pool, "shoulders", "arms", "triceps", "biceps")
	case "lower":
		addCompound(&chosen, pool, "legs")
		addAny(&chosen, pool, "glutes")
		addAny(&chosen, pool, "core")
	case "push":
		addCompound(&chosen, pool, "chest")
		addAny(&chosen, pool, "shoulders")
		addAny(&chosen, pool, "triceps")
	case "pull":
		addCompound(&chosen, pool, "back")
		addAny(&chosen, pool, "biceps")
		addAny(&chosen, pool, "rear_delts", "back")
	case "legs":
		addCompound(&chosen, pool, "legs")
		addAny(&chosen, pool, "glutes")
		addAny(&chosen, pool, "core")
	default: // full
		addCompound(&chosen, pool, "legs")
		addCompound(&chosen, pool, "chest")
		addCompound(&chosen, pool, "back")
	}
	chosen = dedupKeepMax3(chosen)
	items := make([]SessionItem, 0, len(chosen))
	for _, e := range chosen {
		items = append(items, SessionItem{
			ExerciseID: e.ID,
			Series:     seriesFor(e, goal),
			Reps:       repsFor(e, goal),
		})
	}
	return items
}

func addCompound(dst *[]Exercise, src []Exercise, muscle string) {
	for _, e := range src {
		if strings.EqualFold(e.MuscleGroup, muscle) && (e.Pattern == "compound" || e.Pattern == "multi") {
			*dst = append(*dst, e)
			return
		}
	}
	// fallback: qualquer do grupo
	addAny(dst, src, muscle)
}

func addAny(dst *[]Exercise, src []Exercise, muscles ...string) {
	for _, m := range muscles {
		for _, e := range src {
			if strings.EqualFold(e.MuscleGroup, m) {
				*dst = append(*dst, e)
				return
			}
		}
	}
}

func dedupKeepMax3(in []Exercise) []Exercise {
	seen := map[string]bool{}
	out := make([]Exercise, 0, 3)
	for _, e := range in {
		if e.ID == "" || seen[strings.ToLower(e.ID)] {
			continue
		}
		seen[strings.ToLower(e.ID)] = true
		out = append(out, e)
		if len(out) == 3 {
			break
		}
	}
	return out
}

func seriesFor(_ Exercise, goal string) int {
	switch strings.ToLower(goal) {
	case "strength":
		return 5
	case "endurance":
		return 2
	default: // hypertrophy/default
		return 3
	}
}

func repsFor(_ Exercise, goal string) string {
	switch strings.ToLower(goal) {
	case "strength":
		return "3-5"
	case "endurance":
		return "15-20"
	default:
		return "8-12"
	}
}

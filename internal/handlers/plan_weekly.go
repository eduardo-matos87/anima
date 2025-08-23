package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WeeklyPlanDay struct {
	DayIndex   int                 `json:"day_index"`
	Divisao    string              `json:"divisao"`
	TreinoID   string              `json:"treino_id"`
	Exercicios []GeneratedExercise `json:"exercicios"`
	CoachNotes string              `json:"coach_notes,omitempty"`
}

type WeeklyPlanResp struct {
	Objetivo string          `json:"objetivo"`
	Nivel    string          `json:"nivel"`
	Dias     int             `json:"dias"`
	BaseDiv  string          `json:"divisao_base"`
	Items    []WeeklyPlanDay `json:"items"`
}

func PlanWeekly(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Query params
		obj := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("objetivo")))
		niv := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("nivel")))
		div := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("divisao")))
		if obj == "" {
			obj = "hipertrofia"
		}
		if niv == "" {
			niv = "iniciante"
		}
		if div == "" {
			div = "fullbody"
		}

		days := 3
		if v := strings.TrimSpace(r.URL.Query().Get("days")); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 7 {
				days = n
			}
		}

		uid := getUserID(r)
		prof, _ := loadUserProfile(r.Context(), db, uid)
		if wkg, _ := latestWeight(r.Context(), db, uid); wkg != nil {
			prof.WeightKG = wkg
		}

		seq := divisionSequence(div) // sequência de divisões nos dias
		out := make([]WeeklyPlanDay, 0, days)

		for i := 0; i < days; i++ {
			dayDiv := seq[i%len(seq)]
			req := GenerateReq{
				Objetivo: obj,
				Nivel:    niv,
				Divisao:  dayDiv,
				Dias:     days,
				Persist:  ptrBool(false), // preview, não persiste
				TreinoID: "week-" + time.Now().Format("20060102") + "-d" + strconv.Itoa(i+1),
			}
			// monta plano v1.1
			exs, err := buildPlanV11(r.Context(), db, req)
			if err != nil {
				http.Error(w, "erro no planner: "+err.Error(), http.StatusInternalServerError)
				return
			}
			coach := ""
			if prof.UseAI == nil || *prof.UseAI {
				coach = buildCoachNotes(req, prof)
			}
			out = append(out, WeeklyPlanDay{
				DayIndex:   i + 1,
				Divisao:    dayDiv,
				TreinoID:   req.TreinoID,
				Exercicios: exs,
				CoachNotes: coach,
			})
		}

		resp := WeeklyPlanResp{
			Objetivo: obj,
			Nivel:    niv,
			Dias:     days,
			BaseDiv:  div,
			Items:    out,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// Mapeia a divisão base para uma sequência de dias
func divisionSequence(base string) []string {
	b := strings.ToLower(strings.TrimSpace(base))
	switch b {
	case "upperlower":
		return []string{"upper", "lower"}
	case "upper":
		return []string{"upper"}
	case "lower":
		return []string{"lower"}
	case "ppl":
		return []string{"push", "pull", "legs"}
	case "push":
		return []string{"push"}
	case "pull":
		return []string{"pull"}
	case "legs":
		return []string{"legs"}
	default: // fullbody
		return []string{"fullbody"}
	}
}

func ptrBool(b bool) *bool { return &b }

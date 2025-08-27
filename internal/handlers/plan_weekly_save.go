package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WeeklySaveReq struct {
	Objetivo       string `json:"objetivo"`                   // ex: "hipertrofia"
	Nivel          string `json:"nivel"`                      // ex: "iniciante"
	Divisao        string `json:"divisao"`                    // ex: "fullbody" | "upperlower" | "ppl" | "push" | "pull" | "legs"
	Dias           int    `json:"dias"`                       // 1..7 (default 3)
	TreinoIDPrefix string `json:"treino_id_prefix,omitempty"` // ex: "week-20250822"
}

type WeeklySaveItem struct {
	DayIndex int    `json:"day_index"`
	ID       int    `json:"id"`
	TreinoID string `json:"treino_id"`
}

type WeeklySaveResp struct {
	Objetivo string           `json:"objetivo"`
	Nivel    string           `json:"nivel"`
	Dias     int              `json:"dias"`
	BaseDiv  string           `json:"divisao_base"`
	Items    []WeeklySaveItem `json:"items"`
}

func PlanWeeklySave(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req WeeklySaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json inválido", http.StatusBadRequest)
			return
		}

		obj := strings.TrimSpace(strings.ToLower(req.Objetivo))
		niv := strings.TrimSpace(strings.ToLower(req.Nivel))
		div := strings.TrimSpace(strings.ToLower(req.Divisao))
		if obj == "" {
			obj = "hipertrofia"
		}
		if niv == "" {
			niv = "iniciante"
		}
		if div == "" {
			div = "fullbody"
		}
		days := req.Dias
		if days <= 0 || days > 7 {
			days = 3
		}

		prefix := strings.TrimSpace(req.TreinoIDPrefix)
		if prefix == "" {
			prefix = "week-" + time.Now().Format("20060102")
		}

		uid := getUserID(r)
		prof, _ := loadUserProfile(r.Context(), db, uid)
		if wkg, _ := latestWeight(r.Context(), db, uid); wkg != nil {
			prof.WeightKG = wkg
		}

		seq := divisionSequence(div)
		items := make([]WeeklySaveItem, 0, days)

		for i := 0; i < days; i++ {
			dayDiv := seq[i%len(seq)]
			key := prefix + "-d" + strconv.Itoa(i+1)

			genReq := GenerateReq{
				Objetivo: obj,
				Nivel:    niv,
				Divisao:  dayDiv,
				Dias:     days,
				Persist:  ptrBool(true), // salvando
				TreinoID: key,
			}

			// monta plano (usa v1.1 com descanso)
			exs, err := buildPlanV11(r.Context(), db, genReq)
			if err != nil {
				http.Error(w, "erro no planner: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if len(exs) == 0 {
				http.Error(w, "catálogo vazio para gerar plano", http.StatusConflict)
				return
			}

			coach := ""
			if prof.UseAI == nil || *prof.UseAI {
				coach = buildCoachNotes(genReq, prof)
			}

			id, err := persistPlanHandleDup(r.Context(), db, key, genReq, coach, exs)
			if err != nil {
				http.Error(w, "falha ao salvar dia "+strconv.Itoa(i+1)+": "+err.Error(), http.StatusInternalServerError)
				return
			}

			items = append(items, WeeklySaveItem{
				DayIndex: i + 1,
				ID:       id,
				TreinoID: key,
			})
		}

		resp := WeeklySaveResp{
			Objetivo: obj,
			Nivel:    niv,
			Dias:     days,
			BaseDiv:  div,
			Items:    items,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func persistPlanHandleDup(ctx context.Context, db *sql.DB, key string, req GenerateReq, coach string, plan []GeneratedExercise) (int, error) {
	id, err := persistPlan(ctx, db, key, req, coach, plan)
	if err == nil {
		return id, nil
	}
	// se conflitar por chave lógica, tenta variar
	if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
		key2 := key + "-" + time.Now().Format("150405.000")
		return persistPlan(ctx, db, key2, req, coach, plan)
	}
	return 0, err
}

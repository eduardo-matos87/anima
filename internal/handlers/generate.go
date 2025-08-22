package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

type GenerateReq struct {
	Objetivo string `json:"objetivo"`
	Nivel    string `json:"nivel"`
	Divisao  string `json:"divisao"`
	Dias     int    `json:"dias,omitempty"`
}

type GeneratedExercise struct {
	Nome        string `json:"nome"`
	Series      int    `json:"series"`
	Repeticoes  string `json:"repeticoes"`
	DescansoSeg int    `json:"descanso_seg"`
}

type GenerateResp struct {
	TreinoID   string              `json:"treino_id"`
	Exercicios []GeneratedExercise `json:"exercicios"`
	CoachNotes string              `json:"coach_notes,omitempty"`
}

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
		if req.Objetivo == "" || req.Nivel == "" || req.Divisao == "" {
			http.Error(w, "campos obrigatórios: objetivo, nivel, divisao", http.StatusBadRequest)
			return
		}
		if req.Dias <= 0 {
			req.Dias = 3
		}

		uid := getUserID(r)

		// Perfil + última métrica
		prof, _ := loadUserProfile(r.Context(), db, uid)
		if wkg, _ := latestWeight(r.Context(), db, uid); wkg != nil {
			prof.WeightKG = wkg
		}

		// Exercícios mock
		exs := []GeneratedExercise{
			{Nome: "Agachamento Livre", Series: 3, Repeticoes: "8-12", DescansoSeg: 60},
			{Nome: "Supino Reto", Series: 4, Repeticoes: "8-12", DescansoSeg: 90},
			{Nome: "Levantamento Terra", Series: 3, Repeticoes: "8-12", DescansoSeg: 120},
			{Nome: "Remada Curvada", Series: 4, Repeticoes: "8-12", DescansoSeg: 60},
			{Nome: "Desenvolvimento Militar", Series: 3, Repeticoes: "8-12", DescansoSeg: 90},
			{Nome: "Puxada na Frente", Series: 4, Repeticoes: "8-12", DescansoSeg: 120},
		}

		// Coach notes (gera apenas quando use_ai=true)
		coach := ""
		if prof.UseAI != nil && *prof.UseAI {
			coach = buildCoachNotes(req, prof)
		}

		resp := GenerateResp{
			TreinoID:   time.Now().Format("20060102T150405"),
			Exercicios: exs,
			CoachNotes: coach,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// ===== Coach notes =====

func buildCoachNotes(req GenerateReq, p userProfile) string {
	obj := req.Objetivo
	if obj == "" && p.TrainingGoal != nil && *p.TrainingGoal != "" {
		obj = *p.TrainingGoal
	}

	// idade
	var idade *int
	if p.BirthDate != nil {
		y := computeAge(*p.BirthDate, time.Now())
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
		txt += "Dê ênfase à densidade do treino e controle de descanso; mantenha leve déficit calórico."
	case "resistência":
		txt += "Volume moderado e descanso curto; priorize constância e cadência controlada."
	default:
		txt += "Mantenha técnica perfeita, aquecimento e progressão gradual."
	}
	if idade != nil && *idade >= 40 {
		txt += " Aqueça bem ombros/quadril; evite picos de carga abruptos."
	}
	return txt
}

func valOr(def, v string) string {
	if v == "" {
		return def
	}
	return v
}

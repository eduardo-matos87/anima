package ai

import (
	"fmt"
	"math"
	"time"
)

type Profile struct {
	HeightCm      int
	WeightKg      float64
	Birthdate     *time.Time
	Goal          string
	ActivityLevel string
	UseAI         bool
}

// Idade aproximada em anos
func age(bd *time.Time) int {
	if bd == nil {
		return 0
	}
	now := time.Now()
	y := now.Year() - bd.Year()
	// Ajuste se ainda não fez aniversário no ano
	if now.YearDay() < bd.YearDay() {
		y--
	}
	if y < 0 {
		return 0
	}
	return y
}

// IMC simples
func bmi(heightCm int, weightKg float64) float64 {
	if heightCm <= 0 || weightKg <= 0 {
		return 0
	}
	h := float64(heightCm) / 100.0
	return weightKg / (h * h)
}

// Heurística simples para volume e foco com base em objetivo/IMC
func coachNotes(p Profile) string {
	imc := bmi(p.HeightCm, p.WeightKg)
	a := age(p.Birthdate)

	var foco string
	switch p.Goal {
	case "hipertrofia":
		foco = "priorize progressão de carga e boa técnica em compostos"
	case "emagrecimento":
		foco = "mantenha maior densidade do treino e descanso controlado"
	case "resistência":
		foco = "ênfase em volume moderado e menor descanso"
	default:
		foco = "mantenha técnica e regularidade"
	}

	var extra string
	switch {
	case imc == 0:
		extra = "registre peso/altura para recomendações melhores"
	case imc < 18.5:
		extra = "reforço calórico leve pode ajudar no ganho de massa"
	case imc < 25:
		extra = "mantenha leve superávit calórico e sono adequado"
	case imc < 30:
		extra = "combine treino com leve déficit calórico e alta consistência"
	default:
		extra = "priorize aderência e monitoramento semanal de medidas"
	}

	// Toques contextuais (totalmente heurísticos)
	hint := ""
	if a > 0 {
		if a >= 40 {
			hint = "; aqueça bem ombros/quadril e respeite cadências seguras"
		} else if a <= 20 {
			hint = "; foque em técnica antes de buscar cargas muito altas"
		}
	}

	// Arredonda IMC para duas casas para exibir
	roundIMC := math.Round(imc*100) / 100

	return fmt.Sprintf("%s. IMC=%.2f — %s%s.", foco, roundIMC, extra, hint)
}

// ComputeCoachNotes cria uma nota final, ou string vazia se UseAI=false.
func ComputeCoachNotes(p Profile) string {
	if !p.UseAI {
		return ""
	}
	return coachNotes(p)
}

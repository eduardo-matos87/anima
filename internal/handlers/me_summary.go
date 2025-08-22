package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"time"
)

type MeSummary struct {
	HeightCM          *int     `json:"height_cm,omitempty"`
	WeightKG          *float64 `json:"weight_kg,omitempty"` // último peso
	AgeYears          *int     `json:"age_years,omitempty"`
	BMI               *float64 `json:"bmi,omitempty"`
	WeightChange30dKg *float64 `json:"weight_change_30d,omitempty"`   // delta (kg) últimos 30 dias
	LastMeasurementAt *string  `json:"last_measurement_at,omitempty"` // YYYY-MM-DD
}

func MeSummaryHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := getUserID(r)

		// 1) Perfil (altura, peso, birth_date)
		prof, _ := loadUserProfile(r.Context(), db, uid)

		// 2) Métricas (último peso & peso de ~30d atrás)
		now := time.Now().UTC()
		thirtyDaysAgo := now.AddDate(0, 0, -30)

		var lastAt *time.Time
		lastW, lastAtOut, _ := weightAtOrLatest(r.Context(), db, uid, nil) // último
		if lastAtOut != nil {
			lastAt = lastAtOut
		}
		w30, _, _ := weightAtOrLatest(r.Context(), db, uid, &thirtyDaysAgo) // valor mais próximo de 30d atrás

		// Merge peso: prioridade ao último de métricas; senão, do perfil
		weight := prof.WeightKG
		if lastW != nil {
			weight = lastW
		}

		// Idade
		var age *int
		if prof.BirthDate != nil {
			y := computeAge(*prof.BirthDate, now)
			age = &y
		}

		// BMI
		var bmi *float64
		if prof.HeightCM != nil && *prof.HeightCM > 0 && weight != nil && *weight > 0 {
			hm := float64(*prof.HeightCM) / 100.0
			v := *weight / (hm * hm)
			v = math.Round(v*100) / 100
			bmi = &v
		}

		// Delta 30d
		var delta *float64
		if weight != nil && w30 != nil {
			d := *weight - *w30
			d = math.Round(d*100) / 100
			delta = &d
		}

		var lastAtStr *string
		if lastAt != nil {
			s := lastAt.Format("2006-01-02")
			lastAtStr = &s
		}

		out := MeSummary{
			HeightCM:          prof.HeightCM,
			WeightKG:          weight,
			AgeYears:          age,
			BMI:               bmi,
			WeightChange30dKg: delta,
			LastMeasurementAt: lastAtStr,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})
}

// Retorna peso e data da medição mais recente; se refDate != nil,
// tenta pegar a medição mais próxima ANTES ou NA data refDate.
func weightAtOrLatest(ctx context.Context, db *sql.DB, userID string, refDate *time.Time) (*float64, *time.Time, error) {
	if refDate == nil {
		var w sql.NullFloat64
		var at time.Time
		err := db.QueryRowContext(ctx, `
			SELECT weight_kg, measured_at
			FROM user_metrics
			WHERE user_id=$1
			ORDER BY measured_at DESC
			LIMIT 1
		`, userID).Scan(&w, &at)
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		if w.Valid {
			v := w.Float64
			return &v, &at, nil
		}
		return nil, &at, nil
	}

	var w sql.NullFloat64
	var at time.Time
	err := db.QueryRowContext(ctx, `
		SELECT weight_kg, measured_at
		FROM user_metrics
		WHERE user_id=$1 AND measured_at<= $2
		ORDER BY measured_at DESC
		LIMIT 1
	`, userID, *refDate).Scan(&w, &at)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if w.Valid {
		v := w.Float64
		return &v, &at, nil
	}
	return nil, &at, nil
}

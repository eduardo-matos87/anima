package handlers

import (
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
	WeightChange30dKg *float64 `json:"weight_change_30d,omitempty"`   // delta kg
	LastMeasurementAt *string  `json:"last_measurement_at,omitempty"` // YYYY-MM-DD
}

func MeSummaryHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := getUserID(r)

		prof, _ := loadUserProfile(r.Context(), db, uid)

		now := time.Now().UTC()
		thirtyDaysAgo := now.AddDate(0, 0, -30)

		var lastAt *time.Time
		lastW, lastAtOut, _ := weightAtOrLatest(r.Context(), db, uid, nil)
		if lastAtOut != nil {
			lastAt = lastAtOut
		}
		w30, _, _ := weightAtOrLatest(r.Context(), db, uid, &thirtyDaysAgo)

		// peso final (métrica > perfil)
		weight := prof.WeightKG
		if lastW != nil {
			weight = lastW
		}

		// idade
		var age *int
		if prof.BirthDate != nil {
			y := computeAge(*prof.BirthDate, now)
			age = &y
		}

		// IMC
		var bmi *float64
		if prof.HeightCM != nil && *prof.HeightCM > 0 && weight != nil && *weight > 0 {
			hm := float64(*prof.HeightCM) / 100.0
			v := *weight / (hm * hm)
			v = math.Round(v*100) / 100
			bmi = &v
		}

		// delta 30d
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

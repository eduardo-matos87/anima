package handlers

import (
	"context"
	"database/sql"
	"time"
)

// Perfil unificado para uso interno nos handlers
type userProfile struct {
	HeightCM        *int
	WeightKG        *float64
	BirthDate       *time.Time
	TrainingGoal    *string
	ExperienceLevel *string
	ActivityLevel   *string
	UseAI           *bool
}

// Carrega perfil por user_id
func loadUserProfile(ctx context.Context, db *sql.DB, userID string) (userProfile, error) {
	var p userProfile
	var h sql.NullInt64
	var w sql.NullFloat64
	var bd sql.NullTime
	var goal, lvl, act sql.NullString
	var use sql.NullBool

	err := db.QueryRowContext(ctx, `
		SELECT height_cm, weight_kg, birth_date, training_goal, experience_level, activity_level, use_ai
		FROM user_profiles
		WHERE user_id = $1
	`, userID).Scan(&h, &w, &bd, &goal, &lvl, &act, &use)

	if err == nil {
		if h.Valid {
			v := int(h.Int64)
			p.HeightCM = &v
		}
		if w.Valid {
			v := w.Float64
			p.WeightKG = &v
		}
		if bd.Valid {
			t := bd.Time
			p.BirthDate = &t
		}
		if goal.Valid {
			s := goal.String
			p.TrainingGoal = &s
		}
		if lvl.Valid {
			s := lvl.String
			p.ExperienceLevel = &s
		}
		if act.Valid {
			s := act.String
			p.ActivityLevel = &s
		}
		if use.Valid {
			v := use.Bool
			p.UseAI = &v
		}
	}
	return p, err
}

// Último peso do usuário
func latestWeight(ctx context.Context, db *sql.DB, userID string) (*float64, error) {
	var w sql.NullFloat64
	err := db.QueryRowContext(ctx, `
		SELECT weight_kg
		FROM user_metrics
		WHERE user_id = $1
		ORDER BY measured_at DESC
		LIMIT 1
	`, userID).Scan(&w)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if w.Valid {
		v := w.Float64
		return &v, nil
	}
	return nil, nil
}

// Peso mais próximo até uma data de referência (ou o último, se refDate=nil)
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

// Idade em anos (seguro para datas futuras)
func computeAge(birth time.Time, now time.Time) int {
	y := now.Year() - birth.Year()
	anniv := time.Date(now.Year(), birth.Month(), birth.Day(), 0, 0, 0, 0, time.UTC)
	if now.Before(anniv) {
		y--
	}
	if y < 0 {
		return 0
	}
	return y
}

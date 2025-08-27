package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
)

// MeProfile: GET (ler) e PATCH (upsert) perfil do usuário atual.
// Requer user_id (via JWT OptionalAuth ou header X-User-ID).
// GET    /api/me/profile
// PATCH  /api/me/profile   {height_cm, weight_kg, birth_year, gender, level, goal}
func MeProfile(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(GetUserID(r))
		if userID == "" {
			http.Error(w, "unauthorized (missing user id)", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			getMeProfile(db, w, r, userID)
		case http.MethodPatch:
			patchMeProfile(db, w, r, userID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

type meProfile struct {
	UserID    string     `json:"user_id"`
	HeightCM  *float64   `json:"height_cm,omitempty"`
	WeightKG  *float64   `json:"weight_kg,omitempty"`
	BirthYear *int       `json:"birth_year,omitempty"`
	Gender    *string    `json:"gender,omitempty"`
	Level     *string    `json:"level,omitempty"`
	Goal      *string    `json:"goal,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func getMeProfile(db *sql.DB, w http.ResponseWriter, r *http.Request, userID string) {
	var out meProfile
	out.UserID = userID

	row := db.QueryRow(`
		SELECT height_cm, weight_kg, birth_year, gender, level, goal, updated_at
		FROM user_profiles WHERE user_id = $1
	`, userID)

	var (
		height, weight      sql.NullFloat64
		birth               sql.NullInt64
		gender, level, goal sql.NullString
		updated             sql.NullTime
	)
	err := row.Scan(&height, &weight, &birth, &gender, &level, &goal, &updated)
	if err == sql.ErrNoRows {
		jsonWrite(w, http.StatusOK, out) // perfil ainda não criado
		return
	}
	if err != nil {
		internalErr(w, err)
		return
	}

	if height.Valid {
		out.HeightCM = &height.Float64
	}
	if weight.Valid {
		out.WeightKG = &weight.Float64
	}
	if birth.Valid {
		v := int(birth.Int64)
		out.BirthYear = &v
	}
	if gender.Valid {
		s := gender.String
		out.Gender = &s
	}
	if level.Valid {
		s := level.String
		out.Level = &s
	}
	if goal.Valid {
		s := goal.String
		out.Goal = &s
	}
	if updated.Valid {
		t := updated.Time
		out.UpdatedAt = &t
	}

	jsonWrite(w, http.StatusOK, out)
}

func patchMeProfile(db *sql.DB, w http.ResponseWriter, r *http.Request, userID string) {
	var in struct {
		HeightCM  *float64 `json:"height_cm"`
		WeightKG  *float64 `json:"weight_kg"`
		BirthYear *int     `json:"birth_year"`
		Gender    *string  `json:"gender"`
		Level     *string  `json:"level"`
		Goal      *string  `json:"goal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}

	// Validações leves (aplicadas somente se o campo veio no payload)
	nowYear := time.Now().Year()
	if in.HeightCM != nil {
		if *in.HeightCM < 50 || *in.HeightCM > 250 {
			badRequest(w, "height_cm out of range (50..250)")
			return
		}
	}
	if in.WeightKG != nil {
		if *in.WeightKG < 20 || *in.WeightKG > 500 {
			badRequest(w, "weight_kg out of range (20..500)")
			return
		}
	}
	if in.BirthYear != nil {
		if *in.BirthYear < 1900 || *in.BirthYear > nowYear {
			badRequest(w, "birth_year out of range (1900..current year)")
			return
		}
	}

	// UPSERT preservando campos não enviados (COALESCE)
	_, err := db.Exec(`
		INSERT INTO user_profiles (user_id, height_cm, weight_kg, birth_year, gender, level, goal, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
		  height_cm = COALESCE(EXCLUDED.height_cm, user_profiles.height_cm),
		  weight_kg = COALESCE(EXCLUDED.weight_kg, user_profiles.weight_kg),
		  birth_year = COALESCE(EXCLUDED.birth_year, user_profiles.birth_year),
		  gender     = COALESCE(EXCLUDED.gender,     user_profiles.gender),
		  level      = COALESCE(EXCLUDED.level,      user_profiles.level),
		  goal       = COALESCE(EXCLUDED.goal,       user_profiles.goal),
		  updated_at = NOW()
	`, userID, in.HeightCM, in.WeightKG, in.BirthYear, in.Gender, in.Level, in.Goal)
	if err != nil {
		// Se algum processo antigo/cliente bypassar o handler, traduz o CHECK do Postgres para 400
		if pqErr, ok := err.(*pq.Error); ok && string(pqErr.Code) == "23514" {
			translateCheckViolation(w, pqErr)
			return
		}
		internalErr(w, err)
		return
	}

	getMeProfile(db, w, r, userID)
}

// traduz 23514 (check_violation) em uma mensagem amigável (400)
func translateCheckViolation(w http.ResponseWriter, pqErr *pq.Error) {
	field := "unknown"
	msg := "validation failed"

	switch pqErr.Constraint {
	case "ck_user_profiles_height_cm":
		field = "height_cm"
		msg = "height_cm out of range (50..250)"
	case "ck_user_profiles_weight_kg":
		field = "weight_kg"
		msg = "weight_kg out of range (20..500)"
	case "ck_user_profiles_birth_year":
		field = "birth_year"
		msg = "birth_year out of range (1900..current year)"
	}

	jsonWrite(w, http.StatusBadRequest, map[string]any{
		"error":   "validation_failed",
		"field":   field,
		"message": msg,
	})
}

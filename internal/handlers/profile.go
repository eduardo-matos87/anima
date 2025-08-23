package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type userProfileOut struct {
	HeightCM        *int     `json:"height_cm,omitempty"`
	WeightKG        *float64 `json:"weight_kg,omitempty"`
	BirthDate       *string  `json:"birth_date,omitempty"` // YYYY-MM-DD
	TrainingGoal    *string  `json:"training_goal,omitempty"`
	ExperienceLevel *string  `json:"experience_level,omitempty"`
	ActivityLevel   *string  `json:"activity_level,omitempty"`
	UseAI           *bool    `json:"use_ai,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
	UpdatedAt       *string  `json:"updated_at,omitempty"` // RFC3339
}

type userProfileIn struct {
	HeightCM        *int     `json:"height_cm"`
	WeightKG        *float64 `json:"weight_kg"`
	BirthDate       *string  `json:"birth_date"` // YYYY-MM-DD
	TrainingGoal    *string  `json:"training_goal"`
	ExperienceLevel *string  `json:"experience_level"`
	ActivityLevel   *string  `json:"activity_level"`
	UseAI           *bool    `json:"use_ai"`
	Notes           *string  `json:"notes"`
}

// GET: /api/me/profile
// PUT: /api/me/profile  (merge + upsert)
func UserProfile(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getProfile(db, w, r)
		case http.MethodPut:
			putProfile(db, w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func getProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	uid := getUserID(r)

	var h sql.NullInt64
	var wkg sql.NullFloat64
	var bd sql.NullTime
	var goal, lvl, notes, act sql.NullString
	var use sql.NullBool
	var upd sql.NullTime

	err := db.QueryRow(`
		SELECT height_cm, weight_kg, birth_date, training_goal, experience_level,
		       activity_level, use_ai, notes, updated_at
		FROM user_profiles WHERE user_id=$1
	`, uid).Scan(&h, &wkg, &bd, &goal, &lvl, &act, &use, &notes, &upd)

	out := userProfileOut{}
	switch err {
	case nil:
		if h.Valid {
			v := int(h.Int64)
			out.HeightCM = &v
		}
		if wkg.Valid {
			v := wkg.Float64
			out.WeightKG = &v
		}
		if bd.Valid {
			s := bd.Time.Format("2006-01-02")
			out.BirthDate = &s
		}
		if goal.Valid {
			s := goal.String
			out.TrainingGoal = &s
		}
		if lvl.Valid {
			s := lvl.String
			out.ExperienceLevel = &s
		}
		if act.Valid {
			s := act.String
			out.ActivityLevel = &s
		}
		if use.Valid {
			v := use.Bool
			out.UseAI = &v
		}
		if notes.Valid {
			s := notes.String
			out.Notes = &s
		}
		if upd.Valid {
			s := upd.Time.UTC().Format(time.RFC3339)
			out.UpdatedAt = &s
		}
	case sql.ErrNoRows:
		// retorna objeto vazio
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func putProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	uid := getUserID(r)
	var in userProfileIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}

	// Carrega atual
	var curH sql.NullInt64
	var curW sql.NullFloat64
	var curBD sql.NullTime
	var curGoal, curLvl, curAct, curNotes sql.NullString
	var curUse sql.NullBool

	err := db.QueryRow(`
		SELECT height_cm, weight_kg, birth_date, training_goal, experience_level,
		       activity_level, use_ai, notes
		FROM user_profiles WHERE user_id=$1
	`, uid).Scan(&curH, &curW, &curBD, &curGoal, &curLvl, &curAct, &curUse, &curNotes)

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Merge
	if in.HeightCM != nil {
		if *in.HeightCM <= 0 {
			curH = sql.NullInt64{}
		} else {
			curH = sql.NullInt64{Int64: int64(*in.HeightCM), Valid: true}
		}
	}
	if in.WeightKG != nil {
		if *in.WeightKG <= 0 {
			curW = sql.NullFloat64{}
		} else {
			curW = sql.NullFloat64{Float64: *in.WeightKG, Valid: true}
		}
	}
	if in.BirthDate != nil {
		if *in.BirthDate == "" {
			curBD = sql.NullTime{}
		} else {
			t, perr := time.Parse("2006-01-02", *in.BirthDate)
			if perr != nil {
				http.Error(w, "birth_date inválido (YYYY-MM-DD)", http.StatusBadRequest)
				return
			}
			curBD = sql.NullTime{Time: t, Valid: true}
		}
	}
	if in.TrainingGoal != nil {
		if *in.TrainingGoal == "" {
			curGoal = sql.NullString{}
		} else {
			curGoal = sql.NullString{String: *in.TrainingGoal, Valid: true}
		}
	}
	if in.ExperienceLevel != nil {
		if *in.ExperienceLevel == "" {
			curLvl = sql.NullString{}
		} else {
			curLvl = sql.NullString{String: *in.ExperienceLevel, Valid: true}
		}
	}
	if in.ActivityLevel != nil {
		if *in.ActivityLevel == "" {
			curAct = sql.NullString{}
		} else {
			curAct = sql.NullString{String: *in.ActivityLevel, Valid: true}
		}
	}
	if in.UseAI != nil {
		curUse = sql.NullBool{Bool: *in.UseAI, Valid: true}
	}
	if in.Notes != nil {
		if *in.Notes == "" {
			curNotes = sql.NullString{}
		} else {
			curNotes = sql.NullString{String: *in.Notes, Valid: true}
		}
	}

	// Upsert
	_, err = db.Exec(`
		INSERT INTO user_profiles
			(user_id, height_cm, weight_kg, birth_date, training_goal, experience_level, activity_level, use_ai, notes, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now())
		ON CONFLICT (user_id) DO UPDATE SET
			height_cm = EXCLUDED.height_cm,
			weight_kg = EXCLUDED.weight_kg,
			birth_date = EXCLUDED.birth_date,
			training_goal = EXCLUDED.training_goal,
			experience_level = EXCLUDED.experience_level,
			activity_level = EXCLUDED.activity_level,
			use_ai = EXCLUDED.use_ai,
			notes = EXCLUDED.notes,
			updated_at = now()
	`, uid, curH, curW, curBD, curGoal, curLvl, curAct, curUse, curNotes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getProfile(db, w, r)
}

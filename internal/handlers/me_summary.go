package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
)

type summaryOut struct {
	UserID string `json:"user_id"`

	Profile struct {
		HeightCM  *float64 `json:"height_cm,omitempty"`
		WeightKG  *float64 `json:"weight_kg,omitempty"`
		BirthYear *int     `json:"birth_year,omitempty"`
		Gender    *string  `json:"gender,omitempty"`
		Level     *string  `json:"level,omitempty"`
		Goal      *string  `json:"goal,omitempty"`
	} `json:"profile"`

	Usage struct {
		SessionsLast7d  int `json:"sessions_last_7d"`
		SessionsLast30d int `json:"sessions_last_30d"`
		SetsCompleted30 int `json:"sets_completed_30d"`
	} `json:"usage"`

	OverloadTop []struct {
		ExercicioID         int64     `json:"exercicio_id"`
		Requests            int       `json:"requests"`
		AvgSuggestedCargaKG float64   `json:"avg_suggested_carga_kg"`
		LastRequestedAt     time.Time `json:"last_requested_at"`
	} `json:"overload_top"`
}

func MeSummaryHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(GetUserID(r))
		if userID == "" {
			http.Error(w, "unauthorized (missing user id)", http.StatusUnauthorized)
			return
		}

		var out summaryOut
		out.UserID = userID

		// Perfil
		{
			row := db.QueryRow(`SELECT height_cm, weight_kg, birth_year, gender, level, goal
				FROM user_profiles WHERE user_id = $1`, userID)
			var hc, wk sql.NullFloat64
			var by sql.NullInt64
			var g, l, goa sql.NullString
			_ = row.Scan(&hc, &wk, &by, &g, &l, &goa)
			if hc.Valid {
				out.Profile.HeightCM = &hc.Float64
			}
			if wk.Valid {
				out.Profile.WeightKG = &wk.Float64
			}
			if by.Valid {
				v := int(by.Int64)
				out.Profile.BirthYear = &v
			}
			if g.Valid {
				s := g.String
				out.Profile.Gender = &s
			}
			if l.Valid {
				s := l.String
				out.Profile.Level = &s
			}
			if goa.Valid {
				s := goa.String
				out.Profile.Goal = &s
			}
		}

		// Janelas
		now := time.Now().UTC()
		from7 := now.AddDate(0, 0, -7)
		from30 := now.AddDate(0, 0, -30)

		// Uso
		{
			row := db.QueryRow(`
SELECT
  COALESCE((SELECT COUNT(*) FROM workout_sessions
            WHERE (user_id IS NULL OR user_id = $1) AND session_at >= $2),0),
  COALESCE((SELECT COUNT(*) FROM workout_sessions
            WHERE (user_id IS NULL OR user_id = $1) AND session_at >= $3),0),
  COALESCE((SELECT COUNT(*) FROM workout_sets s
            JOIN workout_sessions ws ON ws.id = s.session_id
            WHERE s.completed = TRUE
              AND (ws.user_id IS NULL OR ws.user_id = $1)
              AND ws.session_at >= $3),0)
`, userID, from7, from30)
			_ = row.Scan(&out.Usage.SessionsLast7d, &out.Usage.SessionsLast30d, &out.Usage.SetsCompleted30)
		}

		// Overload top últimos 30d
		{
			rows, _ := db.Query(`
SELECT exercicio_id,
       COUNT(*) AS requests,
       AVG(suggested_carga_kg) AS avg_carga,
       MAX(requested_at) AS last_req
FROM overload_suggestions_log
WHERE user_id = $1 AND requested_at >= $2
GROUP BY exercicio_id
ORDER BY requests DESC, last_req DESC
LIMIT 5
`, userID, from30)
			defer rows.Close()
			for rows.Next() {
				var eID int64
				var req int
				var avg float64
				var last time.Time
				_ = rows.Scan(&eID, &req, &avg, &last)
				out.OverloadTop = append(out.OverloadTop, struct {
					ExercicioID         int64     `json:"exercicio_id"`
					Requests            int       `json:"requests"`
					AvgSuggestedCargaKG float64   `json:"avg_suggested_carga_kg"`
					LastRequestedAt     time.Time `json:"last_requested_at"`
				}{ExercicioID: eID, Requests: req, AvgSuggestedCargaKG: avg, LastRequestedAt: last})
			}
		}

		jsonWrite(w, http.StatusOK, out)
	})
}

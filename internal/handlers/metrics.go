package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type metricIn struct {
	MeasuredAt *string  `json:"measured_at"` // YYYY-MM-DD; default: hoje
	WeightKG   *float64 `json:"weight_kg"`
	BodyfatPct *float64 `json:"bodyfat_pct"`
	HeightCM   *int     `json:"height_cm"`
	NeckCM     *float64 `json:"neck_cm"`
	WaistCM    *float64 `json:"waist_cm"`
	HipCM      *float64 `json:"hip_cm"`
	Notes      *string  `json:"notes"`
}

type metricOut struct {
	MeasuredAt string   `json:"measured_at"`
	WeightKG   *float64 `json:"weight_kg,omitempty"`
	BodyfatPct *float64 `json:"bodyfat_pct,omitempty"`
	HeightCM   *int     `json:"height_cm,omitempty"`
	NeckCM     *float64 `json:"neck_cm,omitempty"`
	WaistCM    *float64 `json:"waist_cm,omitempty"`
	HipCM      *float64 `json:"hip_cm,omitempty"`
	Notes      *string  `json:"notes,omitempty"`
}

// GET:  /api/me/metrics?from=YYYY-MM-DD&to=YYYY-MM-DD
// POST: /api/me/metrics
func UserMetrics(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getMetrics(db, w, r)
		case http.MethodPost:
			postMetric(db, w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func getMetrics(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	uid := getUserID(r)
	q := r.URL.Query()
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -90) // padrão: últimos 90 dias
	to := now

	if s := q.Get("from"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			from = t
		} else {
			http.Error(w, "from inválido (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	}
	if s := q.Get("to"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			to = t
		} else {
			http.Error(w, "to inválido (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	}

	rows, err := db.Query(`
		SELECT measured_at, weight_kg, bodyfat_pct, height_cm, neck_cm, waist_cm, hip_cm, notes
		FROM user_metrics
		WHERE user_id = $1 AND measured_at BETWEEN $2 AND $3
		ORDER BY measured_at
	`, uid, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []metricOut
	for rows.Next() {
		var d metricOut
		var mAt time.Time
		var wkg, bfp, neck, waist, hip sql.NullFloat64
		var h sql.NullInt64
		var notes sql.NullString

		if err := rows.Scan(&mAt, &wkg, &bfp, &h, &neck, &waist, &hip, &notes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		d.MeasuredAt = mAt.Format("2006-01-02")
		if wkg.Valid {
			v := wkg.Float64
			d.WeightKG = &v
		}
		if bfp.Valid {
			v := bfp.Float64
			d.BodyfatPct = &v
		}
		if h.Valid {
			v := int(h.Int64)
			d.HeightCM = &v
		}
		if neck.Valid {
			v := neck.Float64
			d.NeckCM = &v
		}
		if waist.Valid {
			v := waist.Float64
			d.WaistCM = &v
		}
		if hip.Valid {
			v := hip.Float64
			d.HipCM = &v
		}
		if notes.Valid {
			v := notes.String
			d.Notes = &v
		}

		out = append(out, d)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func postMetric(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	uid := getUserID(r)
	var in metricIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}

	// measured_at default = hoje (UTC)
	var mAt time.Time
	if in.MeasuredAt == nil || *in.MeasuredAt == "" {
		mAt = time.Now().UTC()
	} else {
		t, err := time.Parse("2006-01-02", *in.MeasuredAt)
		if err != nil {
			http.Error(w, "measured_at inválido (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		mAt = t
	}

	// Converte para Null*
	var wkg, bfp, neck, waist, hip sql.NullFloat64
	var h sql.NullInt64
	var notes sql.NullString

	if in.WeightKG != nil {
		wkg = sql.NullFloat64{Float64: *in.WeightKG, Valid: true}
	}
	if in.BodyfatPct != nil {
		bfp = sql.NullFloat64{Float64: *in.BodyfatPct, Valid: true}
	}
	if in.HeightCM != nil {
		h = sql.NullInt64{Int64: int64(*in.HeightCM), Valid: true}
	}
	if in.NeckCM != nil {
		neck = sql.NullFloat64{Float64: *in.NeckCM, Valid: true}
	}
	if in.WaistCM != nil {
		waist = sql.NullFloat64{Float64: *in.WaistCM, Valid: true}
	}
	if in.HipCM != nil {
		hip = sql.NullFloat64{Float64: *in.HipCM, Valid: true}
	}
	if in.Notes != nil {
		notes = sql.NullString{String: *in.Notes, Valid: true}
	}

	_, err := db.Exec(`
		INSERT INTO user_metrics
			(user_id, measured_at, weight_kg, bodyfat_pct, height_cm, neck_cm, waist_cm, hip_cm, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (user_id, measured_at) DO UPDATE SET
			weight_kg = EXCLUDED.weight_kg,
			bodyfat_pct = EXCLUDED.bodyfat_pct,
			height_cm = EXCLUDED.height_cm,
			neck_cm = EXCLUDED.neck_cm,
			waist_cm = EXCLUDED.waist_cm,
			hip_cm = EXCLUDED.hip_cm,
			notes = EXCLUDED.notes
	`, uid, mAt, wkg, bfp, h, neck, waist, hip, notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

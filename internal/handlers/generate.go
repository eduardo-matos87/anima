package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type generateInput struct {
	Goal           string   `json:"goal"`
	Split          string   `json:"split"`
	AvailableDays  int      `json:"available_days"`
	Experience     string   `json:"experience"`
	Injuries       []string `json:"injuries"`
	Equipment      []string `json:"equipment_allowed"`
	SessionTimeMin int      `json:"session_time_min"`
}

// Aceita *sql.DB (compatível com seu dbpkg.NewPool)
func GenerateTreino(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var in generateInput
		_ = json.NewDecoder(r.Body).Decode(&in)

		resp := map[string]any{
			"title": "Semana 1 (stub)",
			"days": []map[string]any{
				{"name": "Dia A", "exercises": []any{}},
				{"name": "Dia B", "exercises": []any{}},
			},
			"used_db": db != nil,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"workout": resp})
	}
}

//go:build ignore
// +build ignore

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func TreinosItem(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// path: /api/treinos/{id}
		idStr := strings.TrimPrefix(r.URL.Path, "/api/treinos/")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			// Reusa o handler existente de GET by ID
			GetTreinoByID(db).ServeHTTP(w, r)
		case http.MethodPatch:
			patchTreinoNotes(db, w, r, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

type patchNotesIn struct {
	CoachNotes *string `json:"coach_notes"`
}

type patchNotesOut struct {
	ID         int     `json:"id"`
	CoachNotes *string `json:"coach_notes,omitempty"`
}

func patchTreinoNotes(db *sql.DB, w http.ResponseWriter, r *http.Request, id int) {
	var in patchNotesIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	// pode aceitar null/"" para limpar
	_, err := db.Exec(`UPDATE treinos SET coach_notes = $1 WHERE id = $2`, in.CoachNotes, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := patchNotesOut{ID: id, CoachNotes: in.CoachNotes}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

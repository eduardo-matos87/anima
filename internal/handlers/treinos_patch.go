package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type treinoPatch struct {
	CoachNotes *string `json:"coach_notes,omitempty"`
	Nivel      *string `json:"nivel,omitempty"`
	Objetivo   *string `json:"objetivo,omitempty"`
	Divisao    *string `json:"divisao,omitempty"`
	Dias       *int    `json:"dias,omitempty"`
	TreinoKey  *string `json:"treino_key,omitempty"`
}

// PATCH /api/treinos/{id}
func TreinosUpdate(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			badRequest(w, "missing id")
			return
		}
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || id <= 0 {
			badRequest(w, "invalid id")
			return
		}

		var in treinoPatch
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			badRequest(w, "invalid json")
			return
		}

		q := "UPDATE treinos SET "
		args := []any{}
		i := 1
		setCount := 0

		if in.CoachNotes != nil {
			q += "coach_notes = $" + fmtInt(i) + ", "
			args = append(args, *in.CoachNotes)
			i++
			setCount++
		}
		if in.Nivel != nil {
			q += "nivel = $" + fmtInt(i) + ", "
			args = append(args, *in.Nivel)
			i++
			setCount++
		}
		if in.Objetivo != nil {
			q += "objetivo = $" + fmtInt(i) + ", "
			args = append(args, *in.Objetivo)
			i++
			setCount++
		}
		if in.Divisao != nil {
			q += "divisao = $" + fmtInt(i) + ", "
			args = append(args, *in.Divisao)
			i++
			setCount++
		}
		if in.Dias != nil {
			if *in.Dias < 1 || *in.Dias > 7 {
				badRequest(w, "dias must be between 1 and 7")
				return
			}
			q += "dias = $" + fmtInt(i) + ", "
			args = append(args, *in.Dias)
			i++
			setCount++
		}
		if in.TreinoKey != nil {
			q += "treino_key = $" + fmtInt(i) + ", "
			args = append(args, *in.TreinoKey)
			i++
			setCount++
		}

		if setCount == 0 {
			badRequest(w, "no fields to update")
			return
		}

		// remove vírgula final
		q = strings.TrimSuffix(q, ", ")
		q += " WHERE id = $" + fmtInt(i)
		args = append(args, id)

		if _, err := db.Exec(q, args...); err != nil {
			internalErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

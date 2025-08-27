package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
)

type TreinoListItem struct {
	ID         int64   `json:"id"`
	Nivel      *string `json:"nivel,omitempty"`
	Objetivo   *string `json:"objetivo,omitempty"`
	Divisao    *string `json:"divisao,omitempty"`
	Dias       *int    `json:"dias,omitempty"`
	CoachNotes *string `json:"coach_notes,omitempty"`
	TreinoKey  *string `json:"treino_key,omitempty"`
}

// GET /api/treinos?page=&page_size=&nivel=&objetivo=&divisao=
func TreinosCollection(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		// paginação com limites (usa clampInt do util.go)
		page := clampInt(atoiDefault(q.Get("page"), 1), 1, 1_000_000_000)
		pageSize := clampInt(atoiDefault(q.Get("page_size"), 20), 1, 100)
		offset := (page - 1) * pageSize

		// filtros opcionais
		nivel := q.Get("nivel")
		objetivo := q.Get("objetivo")
		divisao := q.Get("divisao")

		where := "WHERE 1=1"
		args := []any{}
		i := 1
		if nivel != "" {
			where += " AND nivel = $" + fmtInt(i)
			args = append(args, nivel)
			i++
		}
		if objetivo != "" {
			where += " AND objetivo = $" + fmtInt(i)
			args = append(args, objetivo)
			i++
		}
		if divisao != "" {
			where += " AND divisao = $" + fmtInt(i)
			args = append(args, divisao)
			i++
		}

		// total hint
		var total int64
		if err := db.QueryRow("SELECT COUNT(*) FROM treinos "+where, args...).Scan(&total); err != nil {
			internalErr(w, err)
			return
		}

		// lista (ordem recente)
		argsList := append(append([]any{}, args...), pageSize, offset)
		rows, err := db.Query(`
			SELECT id, nivel, objetivo, divisao, dias, coach_notes, treino_key
			FROM treinos `+where+`
			ORDER BY id DESC
			LIMIT $`+fmtInt(i)+` OFFSET $`+fmtInt(i+1), argsList...)
		if err != nil {
			internalErr(w, err)
			return
		}
		defer rows.Close()

		items := []TreinoListItem{}
		for rows.Next() {
			var it TreinoListItem
			if err := rows.Scan(&it.ID, &it.Nivel, &it.Objetivo, &it.Divisao, &it.Dias, &it.CoachNotes, &it.TreinoKey); err != nil {
				internalErr(w, err)
				return
			}
			items = append(items, it)
		}

		jsonWrite(w, http.StatusOK, map[string]any{
			"items":      items,
			"page":       page,
			"page_size":  pageSize,
			"total_hint": total,
		})
	})
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

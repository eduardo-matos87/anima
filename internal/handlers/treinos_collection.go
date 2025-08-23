package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type treinoRow struct {
	ID         int     `json:"id"`
	TreinoKey  *string `json:"treino_key,omitempty"`
	Objetivo   string  `json:"objetivo"`
	Nivel      string  `json:"nivel"`
	Dias       int     `json:"dias"`
	Divisao    string  `json:"divisao"`
	CoachNotes *string `json:"coach_notes,omitempty"`
}

type listResp struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total int         `json:"total"`
	Items []treinoRow `json:"items"`
	Query *string     `json:"query,omitempty"`
	Goal  *string     `json:"goal,omitempty"`
	Level *string     `json:"level,omitempty"`
}

// Handler combina GET (lista) e POST (salvar) em /api/treinos
func TreinosCollection(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listTreinos(db, w, r)
		case http.MethodPost:
			// Reusa o handler existente de salvar
			SaveTreino(db).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func listTreinos(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// paginação
	page := tcParseInt(r.URL.Query().Get("page"), 1)
	limit := tcParseInt(r.URL.Query().Get("limit"), 20)
	page = tcClampInt(page, 1, 100000)
	limit = tcClampInt(limit, 1, 100)

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	goal := strings.TrimSpace(r.URL.Query().Get("goal"))
	level := strings.TrimSpace(r.URL.Query().Get("level"))

	// monta filtros dinâmicos
	where := "WHERE 1=1"
	args := []any{}
	argi := 1

	if q != "" {
		where += " AND (unaccent(objetivo) ILIKE unaccent($" + strconv.Itoa(argi) + ") " +
			"OR unaccent(divisao) ILIKE unaccent($" + strconv.Itoa(argi) + ") " +
			"OR unaccent(nivel) ILIKE unaccent($" + strconv.Itoa(argi) + "))"
		args = append(args, "%"+q+"%")
		argi++
	}
	if goal != "" {
		where += " AND objetivo = $" + strconv.Itoa(argi)
		args = append(args, goal)
		argi++
	}
	if level != "" {
		where += " AND nivel = $" + strconv.Itoa(argi)
		args = append(args, level)
		argi++
	}

	// total
	var total int
	if err := db.QueryRow("SELECT COUNT(1) FROM treinos "+where, args...).Scan(&total); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// paginação
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	rows, err := db.Query(`
		SELECT id, treino_key, objetivo, nivel, dias, divisao, coach_notes
		FROM treinos `+where+`
		ORDER BY id DESC
		LIMIT $`+strconv.Itoa(argi)+` OFFSET $`+strconv.Itoa(argi+1), args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]treinoRow, 0, limit)
	for rows.Next() {
		var it treinoRow
		if err := rows.Scan(&it.ID, &it.TreinoKey, &it.Objetivo, &it.Nivel, &it.Dias, &it.Divisao, &it.CoachNotes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		http.Error(w, rows.Err().Error(), http.StatusInternalServerError)
		return
	}

	resp := listResp{
		Page:  page,
		Limit: limit,
		Total: total,
		Items: items,
	}
	if q != "" {
		resp.Query = &q
	}
	if goal != "" {
		resp.Goal = &goal
	}
	if level != "" {
		resp.Level = &level
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// helpers com nomes únicos para evitar conflito
func tcParseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func tcClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

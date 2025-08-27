package handlers

import (
	"database/sql"
<<<<<<< HEAD
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type TreinoListItem struct {
	ID              int64   `json:"id"`
	Objetivo        string  `json:"objetivo"`
	Nivel           string  `json:"nivel"`
	Dias            int     `json:"dias"`
	Divisao         string  `json:"divisao"`
	TreinoKey       *string `json:"treino_key,omitempty"`
	CountExercicios int     `json:"count_exercicios"`
}

type TreinoListResp struct {
	Total int64            `json:"total"`
	Items []TreinoListItem `json:"items"`
}

func TreinosCollection(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listTreinos(db, w, r)
			return
		case http.MethodPost:
			// Reaproveita o handler existente de POST
			SaveTreino(db).ServeHTTP(w, r)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func listTreinos(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := clampInt(parseInt(r.URL.Query().Get("limit"), 20), 1, 200)
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	// WHERE (com alias t.)
	where := ` WHERE 1=1`
	args := []any{}
	// buscamos sem acento em objetivo/nivel/divisao e também em treino_key (normal)
	if q != "" {
		p := "%" + q + "%"
		where += ` AND (
			unaccent(t.objetivo) ILIKE unaccent(` + place(len(args)+1) + `)
			OR unaccent(t.nivel) ILIKE unaccent(` + place(len(args)+2) + `)
			OR unaccent(t.divisao) ILIKE unaccent(` + place(len(args)+3) + `)
			OR t.treino_key ILIKE ` + place(len(args)+4) + `
		)`
		args = append(args, p, p, p, p)
	}

	// total
	var total int64
	if err := db.QueryRow(`SELECT COUNT(*) FROM treinos t`+where, args...).Scan(&total); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// items (com count de exercícios)
	query := `
		SELECT
			t.id, t.objetivo, t.nivel, t.dias, t.divisao, t.treino_key,
			COALESCE(te.cnt, 0) AS count_exercicios
		FROM treinos t
		LEFT JOIN (
			SELECT treino_id, COUNT(*) AS cnt
			FROM treino_exercicios
			GROUP BY treino_id
		) te ON te.treino_id = t.id
	` + where + `
		ORDER BY t.id DESC
		LIMIT ` + strconv.Itoa(limit) + ` OFFSET ` + strconv.Itoa(offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []TreinoListItem
	for rows.Next() {
		var it TreinoListItem
		var key sql.NullString
		if err := rows.Scan(&it.ID, &it.Objetivo, &it.Nivel, &it.Dias, &it.Divisao, &key, &it.CountExercicios); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if key.Valid {
			k := key.String
			it.TreinoKey = &k
		}
		items = append(items, it)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(TreinoListResp{
		Total: total,
		Items: items,
	})
=======
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
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
}

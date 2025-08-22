package handlers

import (
	"database/sql"
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
}

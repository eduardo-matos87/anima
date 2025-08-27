package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

// tenta resolver o ID do treino por treino_key OU por treino_id (compat)
func resolveTreinoID(ctx context.Context, db *sql.DB, key string) (int64, error) {
	var id int64

	// 1) tenta por treino_key (se a coluna existir/estiver preenchida)
	if err := db.QueryRowContext(ctx, `SELECT id FROM treinos WHERE treino_key = $1`, key).Scan(&id); err == nil {
		return id, nil
	} else if err != sql.ErrNoRows {
		// Ignora erros de coluna inexistente etc. e segue para a 2ª tentativa
	}

	// 2) tenta por treino_id (compat com gerador atual)
	if err := db.QueryRowContext(ctx, `SELECT id FROM treinos WHERE treino_id = $1`, key).Scan(&id); err == nil {
		return id, nil
	}

	return 0, sql.ErrNoRows
}

// GET /api/treinos/by-key/{key}
func TreinosGetByKey(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/api/treinos/by-key/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		key := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, prefix))
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		id, err := resolveTreinoID(r.Context(), db, key)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			internalErr(w, err)
			return
		}

		// delega para o handler de item
		r2 := r.Clone(r.Context())
		r2.URL.Path = fmt.Sprintf("/api/treinos/%d", id)
		TreinosItem(db).ServeHTTP(w, r2)
	})
}

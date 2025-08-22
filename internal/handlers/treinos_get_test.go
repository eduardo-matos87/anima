package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetTreinoByID_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	id := int64(42)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT objetivo, nivel, dias, divisao FROM treinos WHERE id=$1",
	)).WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"objetivo", "nivel", "dias", "divisao"}).
			AddRow("hipertrofia", "iniciante", 3, "fullbody"))

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT te.exercicio_id, e.name AS nome, e.muscle_group AS grupo, te.series, te.repeticoes
			FROM treino_exercicios te
			JOIN exercises e ON e.id = te.exercicio_id
			WHERE te.treino_id = $1
			ORDER BY te.id`,
	)).WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"exercicio_id", "nome", "grupo", "series", "repeticoes"}).
			AddRow(1, "Agachamento Livre", "legs", 3, "8-12").
			AddRow(2, "Supino Reto", "chest", 4, "8-10"))

	req := httptest.NewRequest(http.MethodGet, "/api/treinos/42", nil)
	rr := httptest.NewRecorder()

	h := GetTreinoByID(db) // usa o *sql.DB mockado
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	var got TreinoDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if got.ID != id || got.Dias != 3 || len(got.Exercicios) != 2 {
		t.Fatalf("resposta inesperada: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTreinoByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	id := int64(999)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT objetivo, nivel, dias, divisao FROM treinos WHERE id=$1",
	)).WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"objetivo", "nivel", "dias", "divisao"})) // vazio

	req := httptest.NewRequest(http.MethodGet, "/api/treinos/999", nil)
	rr := httptest.NewRecorder()

	h := GetTreinoByID(db)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

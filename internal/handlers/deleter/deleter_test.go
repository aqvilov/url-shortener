package deleter_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"url-shortener/internal/handlers/deleter"
)

// mock
type mockURLDeleter struct {
	deleteFunc func(alias string) error
}

func (m *mockURLDeleter) DeleteUrl(alias string) error {
	return m.deleteFunc(alias)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// вспомогательная функция: поднимает мини-роутер
// и делает запрос на /{alias}
func doRequest(t *testing.T, d deleter.URLDeleter, alias string) *httptest.ResponseRecorder {
	t.Helper()

	handler := deleter.New(newTestLogger(), d)

	r := chi.NewRouter()
	r.Delete("/{alias}", handler)

	req := httptest.NewRequest(http.MethodDelete, "/"+alias, nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	return rec
}

func TestDeleter_Success(t *testing.T) {
	mock := &mockURLDeleter{
		deleteFunc: func(alias string) error { return nil },
	}

	rec := doRequest(t, mock, "google")

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleter_DBError(t *testing.T) {
	mock := &mockURLDeleter{
		deleteFunc: func(alias string) error {
			return errors.New("database error")
		},
	}

	rec := doRequest(t, mock, "google")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleter_EmptyAlias(t *testing.T) {
	mock := &mockURLDeleter{
		deleteFunc: func(alias string) error { return nil },
	}

	handler := deleter.New(newTestLogger(), mock)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleter_CorrectAliasPassedToDB(t *testing.T) {
	var capturedAlias string

	mock := &mockURLDeleter{
		deleteFunc: func(alias string) error {
			capturedAlias = alias
			return nil
		},
	}

	doRequest(t, mock, "myalias")

	assert.Equal(t, "myalias", capturedAlias)
}

package redirect_test

// Юнит тесты для redirect-хендлера (редирект по алиасу)
//
// redirect возвращает HTTP 302 и заголовок Location с оригинальным URL.
// Проверяем: правильный статус, правильный Location, поведение при ошибках.

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"url-shortener/internal/handlers/redirect"
)

type mockURLGetter struct {
	getFunc func(alias string) (string, error)
}

func (m *mockURLGetter) GetUrl(alias string) (string, error) {
	return m.getFunc(alias)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// вспомогательная функция: мини-роутер + GET /{alias}
func doRequest(t *testing.T, g redirect.URLGetter, alias string) *httptest.ResponseRecorder {
	t.Helper()

	handler := redirect.New(newTestLogger(), g)

	r := chi.NewRouter()
	r.Get("/{alias}", handler)

	req := httptest.NewRequest(http.MethodGet, "/"+alias, nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	return rec
}

func TestRedirect_Success(t *testing.T) {
	mock := &mockURLGetter{
		getFunc: func(alias string) (string, error) {
			return "https://google.com", nil
		},
	}

	rec := doRequest(t, mock, "google")

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://google.com", rec.Header().Get("Location"))
}

// 404
func TestRedirect_NotFound(t *testing.T) {
	mock := &mockURLGetter{
		getFunc: func(alias string) (string, error) {
			return "", errors.New("alias not found")
		},
	}

	rec := doRequest(t, mock, "nonexistent")

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// код 400
func TestRedirect_EmptyAlias(t *testing.T) {
	mock := &mockURLGetter{
		getFunc: func(alias string) (string, error) {
			return "https://google.com", nil
		},
	}

	handler := redirect.New(newTestLogger(), mock)

	// без роутера chi.URLParam вернёт ""(поэтому мы всегда проверяли "" в других файлаъх)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

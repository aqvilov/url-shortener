package redirect

import (
	"log/slog"
	"net/http"
	"url-shortener/internal/handlers/api"

	"github.com/go-chi/chi/v5"
)

type URLGetter interface {
	GetUrl(alias string) (string, error)
}

type Response struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status"`
}

func New(log *slog.Logger, getter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := log.With(slog.String("op", "handlers.redirect.New"))

		alias := chi.URLParam(r, "alias") // URLParam находит алиас внутри запроса, которыйр сделал пользователь
		if alias == "" {
			log.Error("Missing alias")

			api.JSONHandler(w, "Error", "Missing alias", http.StatusBadRequest)
			return
		}

		//обращение к бд
		originalURL, err := getter.GetUrl(alias)
		if err != nil {
			log.Error("Error getting originalURL")

			api.JSONHandler(w, "Error", "Short URL not found", http.StatusNotFound)

			return
		}

		log.Info("successful redirect", "alias", alias, "to original-url", originalURL)
		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

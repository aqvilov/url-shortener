package redirect

import (
	"encoding/json"
	"log/slog"
	"net/http"

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

			JSONHandler(w, "Error", "Missing alias", http.StatusBadRequest)
			return
		}

		//обращение к бд
		originalURL, err := getter.GetUrl(alias)
		if err != nil {
			log.Error("Error getting originalURL")

			JSONHandler(w, "Error", "Short URL not found", http.StatusNotFound)

			return
		}

		log.Info("successful redirect", "alias", alias, "to original-url", originalURL)
		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

// на гениалычах вынес
func JSONHandler(w http.ResponseWriter, error string, status string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(Response{
		Error:  error,
		Status: status,
	})
	if err != nil {
		return
	}
}

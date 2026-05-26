package deleter

import (
	"log/slog"
	"net/http"
	"url-shortener/internal/handlers/api"

	"github.com/go-chi/chi/v5"
)

type URLDeleter interface {
	DeleteUrl(alias string) error
}

type Response struct {
	Error error `json:"error,omitempty"`
	// omitempty - если поле осталось пустым, то добавлять его не надо
	Status string `json:"status"`
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := log.With(slog.String("op", "handlers.deleter.New"))

		getAlias := chi.URLParam(r, "alias")
		if getAlias == "" {
			log.Error("Error getting alias")

			api.JSONHandler(w, "Error", "Error getting alias", http.StatusBadRequest)
			return
		}

		err := deleter.DeleteUrl(getAlias)
		if err != nil {
			log.Error("fail to delete url", "err", err)

			api.JSONHandler(w, "Error", "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("url deleted successfully", "alias", getAlias)

	}
}

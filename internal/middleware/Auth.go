package CustomMiddleware

import (
	"log"
	"net/http"
	"os"
	"url-shortener/internal/handlers/api"

	"github.com/joho/godotenv"
)

// если проект будет хостится на сервере, к нему сможет подключиться любой
// тогда нужна авторизация, чтобы удалять ссылки мог только авторизованный

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := godotenv.Load(); err != nil {
			log.Printf(".env не найден для получения middleware токена: %v", err)
		}

		token := os.Getenv("MIDDLEWARE_TOKEN")
		if token == "" {
			api.JSONHandler(w, "didnt get middleware_token", "error", http.StatusInternalServerError)
			return
		}
		// заголовок authorization: Bearer <token>
		header := r.Header.Get("Authorization")
		if header != "Bearer "+token {
			api.JSONHandler(w, "unauthorized", "error", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

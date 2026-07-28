package CustomMiddleware

import (
	"net/http"
	"url-shortener/internal/handlers/api"
)

// если проект будет хостится на сервере, к нему сможет подключиться любой
// тогда нужна авторизация, чтобы удалять ссылки мог только авторизованный

// token читается один раз при старте сервера и захватывается замыканием,
// чтобы не перечитывать .env на каждый запрос
func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// заголовок authorization: Bearer <token>
			header := r.Header.Get("Authorization")
			if header != "Bearer "+token {
				api.JSONHandler(w, "unauthorized", "error", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

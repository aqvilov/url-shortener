package CustomMiddleware

import (
	"net/http"
	"sync"
	"time"
)

type user struct {
	lastVisit time.Time
}

func RateLimit(next http.Handler) http.Handler {
	mutex := sync.Mutex{}

	//храним ip и время активности
	seen := make(map[string]*user)

	go func() {
		for {
			time.Sleep(1 * time.Minute) // проверяем каждую минуту
			mutex.Lock()
			for ip, lastVisit := range seen {
				if time.Since(lastVisit.lastVisit) > (5 * time.Minute) { // если пользователь неактивен больше 5 минут, то удаляем его из мапы
					delete(seen, ip)
				}
			}
			mutex.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ip:port
		ip := r.RemoteAddr

		mutex.Lock()

		u, exists := seen[ip]
		if !exists {
			seen[ip] = &user{lastVisit: time.Now()}
			mutex.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		if time.Since(u.lastVisit) < 500*time.Millisecond {
			u.lastVisit = time.Now()
			mutex.Unlock()

			http.Error(w, "Too many requests.\nError: ", http.StatusTooManyRequests)
			return
		}

		u.lastVisit = time.Now()
		mutex.Unlock()

		next.ServeHTTP(w, r)
	})

}

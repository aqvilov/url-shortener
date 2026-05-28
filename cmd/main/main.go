package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/handlers/alias"
	"url-shortener/internal/handlers/deleter"
	"url-shortener/internal/handlers/redirect"
	"url-shortener/internal/logger"
	CustomMiddleware "url-shortener/internal/middleware"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	//init config
	cfg := config.Load()
	//init logger
	log := logger.SetupLogger(cfg.Env)

	log.Info("run url-shortener", "env", cfg.Env)

	db, err := storage.New(log)
	if err != nil {
		log.Error("failed to initialize storage", "err", err)
		os.Exit(1)
	}

	router := chi.NewRouter()

	//middleware/
	router.Use(CustomMiddleware.NewLogger(log))
	router.Use(CustomMiddleware.RateLimit)

	router.Use(middleware.Recoverer) // чтобы сервер не падал от одной паники

	//подключаем написанные хендлеры к роутеру
	router.Post("/url", alias.New(log, db))
	router.Get("/{alias}", redirect.New(log, db))
	router.Delete("/url/{alias}", deleter.New(log, db))

	log.Info("Start server")

	server := &http.Server{
		Addr:         cfg.HTTPServer.Addr,
		Handler:      router,                 // роутер сам по себе тоже является хендлером
		ReadTimeout:  cfg.HTTPServer.Timeout, // время на то чтобы прочитать запрос
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
	}

	go func() {
		log.Info("Start server", "addr", cfg.HTTPServer.Addr)

		if err := server.ListenAndServe(); err != nil { // тут мы блокируемся и не идем дальше (вынес в горутину)
			log.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	//              os.Interrupt тоже самое что SIGINT (сигнал ctrl + c), SIGTERM - ловит сигнал докера, кубера, системД
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// server.Shutdown ждёт завершения активных запросов (контекст с таймаутом ограничивает время ожидания)
	if err := server.Shutdown(ctx); err != nil {
		// если за 5 секунд запросы не завершились, то shutdown вызывает ошибку
		// логируем эту ошибку
		log.Info("failed to shutdown server", "err", err)
	}
	//end
	log.Info("end url-shortener")
}

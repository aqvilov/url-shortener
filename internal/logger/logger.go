package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	// читаемый хендлер
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // все уровни
		AddSource: true,            // отображать файл и строку
	})

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // все уровни
		AddSource: true,            // отображать файл и строку
	})

	prodJSONHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo, // в проде не нагружать debug-логами
		AddSource: true,
	})

	//в зависимости от выбранного окружения свитчимся по хендлерам
	switch env {
	case "local":
		log = slog.New(textHandler)
	case "test":
		log = slog.New(jsonHandler)
	case "prod":
		log = slog.New(prodJSONHandler)
	default:
		log = slog.New(textHandler)
	}
	return log
}

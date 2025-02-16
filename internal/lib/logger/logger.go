package logger

import (
	"log/slog"
	"os"

	"github.com/fatih/color"
	"github.com/linemk/avito-shop/internal/lib/logger/handlers/slogpretty"
)

// switching logger
const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

// SetupLogger инициализирует логгер в зависимости от переданного окружения
// для локальной разработки используется цветной вывод (pretty), а для dev/prod – JSON
func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case EnvLocal:
		log = setupPrettySlog()
	case EnvDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	color.NoColor = false

	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}

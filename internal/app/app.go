package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
	"github.com/linemk/avito-shop/internal/config"
)

type App struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *sql.DB
}

// NewApp создаёт новый экземпляр App
func NewApp(log *slog.Logger, cfg *config.Config) (*App, error) {

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is not set")
	}
	// реализуем подключение к БД через DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User,
		dbPassword,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	app := &App{
		Config: cfg,
		Logger: log,
		DB:     db,
	}

	return app, nil
}

package main

import (
	"context"

	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/linemk/avito-shop/internal/app"
	"github.com/linemk/avito-shop/internal/app/handlers"
	"github.com/linemk/avito-shop/internal/config"
	"github.com/linemk/avito-shop/internal/jwt-new/jwtmiddleware"
	"github.com/linemk/avito-shop/internal/lib/logger"
	"github.com/linemk/avito-shop/internal/lib/logger/handlers/urllog"
	"github.com/linemk/avito-shop/internal/service"
	"github.com/linemk/avito-shop/internal/storage"
	"github.com/pkg/errors"
)

func main() {
	// загрузка конфигурации
	cfg := config.MustLoad()

	// инициализация логгера, зависит от настройки окружения
	log := logger.SetupLogger(cfg.Env)
	log.Info("starting app", slog.String("env", cfg.Env))

	// загружаем объект приложения, конфигом и подключением к БД
	application, err := app.NewApp(log, cfg)
	if err != nil {
		log.Error("failed to initialize app", slog.Any("error", err))
		panic(errors.Wrap(err, "failed to initialize app"))
	}
	defer application.DB.Close()

	router := chi.NewRouter()
	// настройка middleware
	router.Use(middleware.RequestID)
	router.Use(urllog.CustomLoggerMiddleware(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// реализация слоев по работе с БД по каждому направлению
	userRepo := storage.NewUserRepository(application.DB)
	merchRepo := storage.NewMerchRepository(application.DB)
	orderRepo := storage.NewOrderRepository(application.DB)
	coinTxRepo := storage.NewCoinTransactionRepository(application.DB)

	authService := service.NewAuthService(application.Logger, userRepo, time.Duration(application.Config.JWT.TokenTTL)*time.Minute)
	buyService := service.NewBuyService(application.Logger, application.DB, userRepo, merchRepo, orderRepo)
	sendCoinService := service.NewSendCoinService(application.Logger, application.DB, userRepo, coinTxRepo)
	infoService := service.NewInfoService(application.Logger, userRepo, orderRepo, coinTxRepo)

	// эндпоинт для аутентификации
	router.Post("/api/auth", handlers.AuthHandler(application.Logger, authService))

	router.Group(func(r chi.Router) {
		jwtMW := jwtmiddleware.NewJWTMiddleware()
		r.Use(jwtMW)
		// эндпоинт для инфо
		r.Get("/api/info", handlers.InfoHandler(application.Logger, infoService))
		// эндпоинт для отправки монет другому пользователю
		r.Post("/api/sendCoin", handlers.SendCoinHandler(application.Logger, sendCoinService))
		// эндпоинт для покупки мерча (параметр в path — название товара)
		r.Get("/api/buy/{item}", handlers.BuyHandler(application.Logger, buyService))
	})

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.Any("error", err))
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSign := <-stop
	log.Info("received shutdown signal", slog.String("signal", stopSign.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", slog.Any("error", err))
	}
	log.Info("server gracefully stopped")
}

package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/linemk/avito-shop/internal/storage"
)

type BuyService interface {
	Buy(ctx context.Context, userID int64, item string) error
}

type buyService struct {
	log       *slog.Logger
	userRepo  storage.UserStorage
	merchRepo storage.MerchStorage
	orderRepo storage.OrderStorage
	db        *sql.DB
}

func NewBuyService(log *slog.Logger, db *sql.DB, userRepo storage.UserStorage, merchRepo storage.MerchStorage, orderRepo storage.OrderStorage) BuyService {
	return &buyService{
		log:       log,
		db:        db,
		userRepo:  userRepo,
		merchRepo: merchRepo,
		orderRepo: orderRepo,
	}
}

// Buy осуществляет покупку товара
// Если что-то идет не так, транзакция откатывается
func (s *buyService) Buy(ctx context.Context, userID int64, item string) error {
	const op = "service.BuyService.Buy"
	logger := s.log.With(slog.String("op", op), slog.Int64("userID", userID), slog.String("item", item))
	logger.Info("starting purchase transaction")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("failed to begin transaction", slog.Any("error", err))
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	// Получаем мерч по названию через транзакцию
	merch, err := s.merchRepo.GetMerchByName(ctx, tx, item)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to get merch", slog.Any("error", err))
		return fmt.Errorf("%s: failed to get merch: %w", op, err)
	}

	// Получаем пользователя через транзакцию
	user, err := s.userRepo.LockUserByIDTx(ctx, tx, userID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to get user", slog.Any("error", err))
		return fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	// Проверяем, достаточно ли средств
	if user.CoinBalance < merch.Price {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Warn("insufficient funds", slog.Int("balance", user.CoinBalance), slog.Int("price", merch.Price))
		return fmt.Errorf("%s: insufficient funds", op)
	}

	// Обновляем баланс пользователя
	newBalance := user.CoinBalance - merch.Price
	if err := s.userRepo.UpdateUserBalance(ctx, tx, userID, newBalance); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to update user balance", slog.Any("error", err))
		return fmt.Errorf("%s: failed to update user balance: %w", op, err)
	}

	// Создаем заказ
	if err := s.orderRepo.CreateOrder(ctx, tx, userID, merch.ID, 1, merch.Price); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to create order", slog.Any("error", err))
		return fmt.Errorf("%s: failed to create order: %w", op, err)
	}

	// Коммит транзакции
	if err := tx.Commit(); err != nil {
		logger.Error("failed to commit transaction", slog.Any("error", err))
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	logger.Info("purchase completed successfully")
	return nil
}

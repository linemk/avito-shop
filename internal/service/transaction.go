package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/linemk/avito-shop/internal/storage"
)

// SendCoinService определяет интерфейс для перевода монет.
type SendCoinService interface {
	SendCoin(ctx context.Context, fromUserID int64, toUser string, amount int) error
}

type sendCoinService struct {
	log        *slog.Logger
	db         *sql.DB
	userRepo   storage.UserStorage
	coinTxRepo storage.CoinTransactionStorage
}

func NewSendCoinService(log *slog.Logger, db *sql.DB, userRepo storage.UserStorage, coinTxRepo storage.CoinTransactionStorage) SendCoinService {
	return &sendCoinService{
		log:        log,
		db:         db,
		userRepo:   userRepo,
		coinTxRepo: coinTxRepo,
	}
}

func (s *sendCoinService) SendCoin(ctx context.Context, fromUserID int64, toUser string, amount int) error {
	const op = "service.SendCoinService.SendCoin"
	logger := s.log.With(
		slog.String("op", op),
		slog.Int64("fromUserID", fromUserID),
		slog.String("toUser", toUser),
		slog.Int("amount", amount),
	)
	logger.Info("starting coin transfer transaction")

	if amount <= 0 {
		return fmt.Errorf("%s: amount must be positive", op)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("failed to begin transaction", slog.Any("error", err))
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	// Получаем отправителя через метод LockUserByIDTx (используем транзакцию)
	sender, err := s.userRepo.LockUserByIDTx(ctx, tx, fromUserID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to get sender", slog.Any("error", err))
		return fmt.Errorf("%s: failed to get sender: %w", op, err)
	}

	// Получаем получателя по email (username)
	receiver, err := s.userRepo.GetUserByEmail(ctx, toUser)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			logger.Error("receiver not found", slog.String("toUser", toUser))
			return fmt.Errorf("%s: receiver not found", op)
		}
		logger.Error("failed to get receiver", slog.Any("error", err))
		return fmt.Errorf("%s: failed to get receiver: %w", op, err)
	}

	// проверяем, не отправитель ли пытается сам себе перевести деньги
	if fromUserID == receiver.ID {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("cannot transfer coins to yourself")
		return fmt.Errorf("%s: cannot transfer coins to yourself", op)
	}

	// Проверяем, достаточно ли средств у отправителя
	if sender.CoinBalance < amount {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Warn("insufficient funds", slog.Int("senderBalance", sender.CoinBalance))
		return fmt.Errorf("%s: insufficient funds", op)
	}

	// Обновляем баланс отправителя: списываем монеты
	newSenderBalance := sender.CoinBalance - amount
	if err := s.userRepo.UpdateUserBalance(ctx, tx, fromUserID, newSenderBalance); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to update sender balance", slog.Any("error", err))
		return fmt.Errorf("%s: failed to update sender balance: %w", op, err)
	}

	// Обновляем баланс получателя: прибавляем монеты
	newReceiverBalance := receiver.CoinBalance + amount
	if err := s.userRepo.UpdateUserBalance(ctx, tx, receiver.ID, newReceiverBalance); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to update receiver balance", slog.Any("error", err))
		return fmt.Errorf("%s: failed to update receiver balance: %w", op, err)
	}

	// Регистрируем транзакцию для отправителя (положительная сумма, тип "transfer_sent")
	if err := s.coinTxRepo.CreateTransaction(ctx, tx, fromUserID, amount, "transfer_sent", &receiver.ID); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to record sender transaction", slog.Any("error", err))
		return fmt.Errorf("%s: failed to record sender transaction: %w", op, err)
	}

	// Регистрируем транзакцию для получателя (положительная сумма, тип "transfer_received")
	if err := s.coinTxRepo.CreateTransaction(ctx, tx, receiver.ID, amount, "transfer_received", &fromUserID); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("transaction rollback failed", slog.Any("error", rbErr))
		}
		logger.Error("failed to record receiver transaction", slog.Any("error", err))
		return fmt.Errorf("%s: failed to record receiver transaction: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("failed to commit transaction", slog.Any("error", err))
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	logger.Info("coin transfer completed successfully")
	return nil
}

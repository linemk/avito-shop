package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/linemk/avito-shop/internal/domain/models"
)

// CoinTransactionStorage описывает методы для работы с транзакциями.
type CoinTransactionStorage interface {
	// CreateTransaction создает запись о транзакции.
	CreateTransaction(ctx context.Context, tx *sql.Tx, userID int64, amount int, txType string, relatedUserID *int64) error
	// GetTransactionsByUserID возвращает список транзакций для указанного пользователя.
	GetTransactionsByUserID(ctx context.Context, userID int64) ([]*models.CoinTransaction, error)
}

type coinTransactionRepository struct {
	db *sql.DB
}

func NewCoinTransactionRepository(db *sql.DB) CoinTransactionStorage {
	return &coinTransactionRepository{db: db}
}

func (r *coinTransactionRepository) CreateTransaction(ctx context.Context, tx *sql.Tx, userID int64, amount int, txType string, relatedUserID *int64) error {
	query := `INSERT INTO coin_transactions (user_id, amount, type, related_user_id, created_at)
	          VALUES ($1, $2, $3, $4, NOW())`
	_, err := tx.ExecContext(ctx, query, userID, amount, txType, relatedUserID)
	if err != nil {
		return fmt.Errorf("failed to create coin transaction: %w", err)
	}
	return nil
}

func (r *coinTransactionRepository) GetTransactionsByUserID(ctx context.Context, userID int64) ([]*models.CoinTransaction, error) {
	query := `
		SELECT id, user_id, amount, type, related_user_id, created_at
		FROM coin_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query coin transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.CoinTransaction
	for rows.Next() {
		tx := &models.CoinTransaction{}
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.RelatedUserID, &tx.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan coin transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

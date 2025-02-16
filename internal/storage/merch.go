package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/linemk/avito-shop/internal/domain/models"
)

// MerchStorage описывает методы для работы с таблицей мерча.
type MerchStorage interface {
	// GetMerchByName получает мерч по его названию, используя транзакцию.
	GetMerchByName(ctx context.Context, tx *sql.Tx, name string) (*models.Merch, error)
}

// merchRepository — конкретная реализация интерфейса MerchStorage.
type merchRepository struct {
	db *sql.DB
}

// NewMerchRepository создаёт новый репозиторий мерча.
func NewMerchRepository(db *sql.DB) MerchStorage {
	return &merchRepository{db: db}
}

var ErrMerchNotFound = errors.New("merch not found")

// GetMerchByName ищет мерч по имени в таблице merch.
func (r *merchRepository) GetMerchByName(ctx context.Context, tx *sql.Tx, name string) (*models.Merch, error) {
	merch := &models.Merch{}
	query := "SELECT id, name, price FROM merch WHERE name = $1"
	row := tx.QueryRowContext(ctx, query, name)
	if err := row.Scan(&merch.ID, &merch.Name, &merch.Price); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMerchNotFound
		}
		return nil, err
	}
	return merch, nil
}

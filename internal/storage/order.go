package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/linemk/avito-shop/internal/domain/models"
)

// OrderStorage описывает методы для работы с заказами.
type OrderStorage interface {
	// CreateOrder вставляет новый заказ в таблицу orders с использованием транзакции.
	CreateOrder(ctx context.Context, tx *sql.Tx, userID int64, merchID int64, quantity int, totalPrice int) error
	// GetOrdersByUserID возвращает список заказов для указанного пользователя, с JOIN для получения имени товара.
	GetOrdersByUserID(ctx context.Context, userID int64) ([]*models.Order, error)
}

// orderRepository — конкретная реализация OrderStorage.
type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository создаёт новый репозиторий заказов.
func NewOrderRepository(db *sql.DB) OrderStorage {
	return &orderRepository{db: db}
}

// CreateOrder вставляет новый заказ в таблицу orders.
func (r *orderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, userID int64, merchID int64, quantity int, totalPrice int) error {
	query := `INSERT INTO orders (user_id, merch_id, quantity, total_price, created_at) 
	          VALUES ($1, $2, $3, $4, NOW())`
	_, err := tx.ExecContext(ctx, query, userID, merchID, quantity, totalPrice)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

// GetOrdersByUserID возвращает список заказов для пользователя с JOIN, чтобы получить имя товара.
func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]*models.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.merch_id, m.name, o.quantity, o.total_price, o.created_at
		FROM orders o
		JOIN merch m ON o.merch_id = m.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(&order.ID, &order.UserID, &order.MerchID, &order.MerchName, &order.Quantity, &order.TotalPrice, &order.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

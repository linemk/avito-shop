package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/linemk/avito-shop/internal/domain/models"
)

// Добавим метод GetUserByID в репозиторий.
func (r *userRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	row := r.db.QueryRowContext(ctx, "SELECT id, username, pass_hash, coin_balance FROM users WHERE id = $1", id)
	if err := row.Scan(&user.ID, &user.Email, &user.PassHash, &user.CoinBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

package storage_test

import (
	"context"
	"errors"
	"github.com/linemk/avito-shop/internal/domain/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/linemk/avito-shop/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestGetUserByID_Success(t *testing.T) {
	// Создаем sqlmock для эмуляции базы данных.
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Создаем репозиторий.
	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	userID := int64(1)

	// Подготавливаем ожидаемые строки результата.
	rows := sqlmock.NewRows([]string{"id", "username", "pass_hash", "coin_balance"}).
		AddRow(userID, "test@example.com", []byte("hashed-password"), 1000)

	// Ожидаем выполнение запроса с аргументом userID.
	mock.ExpectQuery("SELECT id, username, pass_hash, coin_balance FROM users WHERE id = \\$1").
		WithArgs(userID).WillReturnRows(rows)

	// Вызываем тестируемую функцию.
	user, err := repo.GetUserByID(ctx, userID)
	assert.NoError(t, err, "Expected no error when user is found")
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, []byte("hashed-password"), user.PassHash)
	assert.Equal(t, 1000, user.CoinBalance)

	// Проверяем, что все ожидания sqlmock выполнены.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetUserByID_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	userID := int64(2)

	// Эмулируем ситуацию, когда запрос возвращает 0 строк.
	rows := sqlmock.NewRows([]string{"id", "username", "pass_hash", "coin_balance"})
	mock.ExpectQuery("SELECT id, username, pass_hash, coin_balance FROM users WHERE id = \\$1").
		WithArgs(userID).WillReturnRows(rows)

	user, err := repo.GetUserByID(ctx, userID)
	assert.Error(t, err, "Expected error when user is not found")
	assert.Nil(t, user, "User should be nil when not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetUserByID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	userID := int64(3)

	// Эмулируем ошибку выполнения запроса.
	mock.ExpectQuery("SELECT id, username, pass_hash, coin_balance FROM users WHERE id = \\$1").
		WithArgs(userID).WillReturnError(errors.New("db error"))

	user, err := repo.GetUserByID(ctx, userID)
	assert.Error(t, err, "Expected error when query fails")
	assert.Nil(t, user, "User should be nil when query fails")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetMerchByName_Success(t *testing.T) {
	// Создаем sqlmock для эмуляции БД.
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем вызов Begin перед тем, как вызвать db.Begin().
	mock.ExpectBegin()
	repo := storage.NewMerchRepository(db)
	ctx := context.Background()
	merchName := "t-shirt"

	tx, err := db.Begin()
	assert.NoError(t, err)

	// Подготавливаем ожидаемые строки результата.
	rows := sqlmock.NewRows([]string{"id", "name", "price"}).
		AddRow(1, merchName, 80)

	// Ожидаем запрос с аргументом merchName.
	query := "SELECT id, name, price FROM merch WHERE name = \\$1"
	mock.ExpectQuery(query).WithArgs(merchName).WillReturnRows(rows)

	// Вызываем GetMerchByName.
	result, err := repo.GetMerchByName(ctx, tx, merchName)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, merchName, result.Name)
	assert.Equal(t, 80, result.Price)

	// Ожидаем вызов Commit и коммитим транзакцию.
	mock.ExpectCommit()
	err = tx.Commit()
	assert.NoError(t, err)

	// Проверяем, что все ожидания выполнены.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetMerchByName_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	repo := storage.NewMerchRepository(db)
	ctx := context.Background()
	merchName := "non-existent"

	tx, err := db.Begin()
	assert.NoError(t, err)

	// Эмулируем ситуацию, когда запрос возвращает 0 строк.
	rows := sqlmock.NewRows([]string{"id", "name", "price"})
	query := "SELECT id, name, price FROM merch WHERE name = \\$1"
	mock.ExpectQuery(query).WithArgs(merchName).WillReturnRows(rows)

	result, err := repo.GetMerchByName(ctx, tx, merchName)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrMerchNotFound))
	assert.Nil(t, result)

	mock.ExpectRollback()
	err = tx.Rollback()
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetMerchByName_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	repo := storage.NewMerchRepository(db)
	ctx := context.Background()
	merchName := "t-shirt"

	tx, err := db.Begin()
	assert.NoError(t, err)

	// Эмулируем ошибку выполнения запроса.
	query := "SELECT id, name, price FROM merch WHERE name = \\$1"
	expectedError := errors.New("query error")
	mock.ExpectQuery(query).WithArgs(merchName).WillReturnError(expectedError)

	result, err := repo.GetMerchByName(ctx, tx, merchName)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)

	mock.ExpectRollback()
	err = tx.Rollback()
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCreateOrder_Success(t *testing.T) {
	// Создаем sqlmock для эмуляции БД.
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewOrderRepository(db)
	ctx := context.Background()

	// Ожидаем вызов BeginTx.
	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)

	// Формируем ожидаемый SQL-запрос, используя regexp.QuoteMeta,
	// чтобы экранировать специальные символы.
	query := regexp.QuoteMeta("INSERT INTO orders (user_id, merch_id, quantity, total_price, created_at) VALUES ($1, $2, $3, $4, NOW())")
	mock.ExpectExec(query).WithArgs(1, 2, 3, 150).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateOrder(ctx, tx, 1, 2, 3, 150)
	assert.NoError(t, err)

	// Ожидаем Commit.
	mock.ExpectCommit()
	err = tx.Commit()
	assert.NoError(t, err)

	// Проверяем, что все ожидания sqlmock выполнены.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetOrdersByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewOrderRepository(db)
	ctx := context.Background()
	userID := int64(1)

	// Подготавливаем ожидаемые строки результата с полями: id, user_id, merch_id, m.name, quantity, total_price, created_at.
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "merch_id", "name", "quantity", "total_price", "created_at"}).
		AddRow(1, userID, 2, "t-shirt", 1, 80, now)
	query := `
		SELECT o\.id, o\.user_id, o\.merch_id, m\.name, o\.quantity, o\.total_price, o\.created_at
		FROM orders o
		JOIN merch m ON o\.merch_id = m\.id
		WHERE o\.user_id = \$1
		ORDER BY o\.created_at DESC`
	mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)

	orders, err := repo.GetOrdersByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, int64(1), orders[0].ID)
	assert.Equal(t, "t-shirt", orders[0].MerchName)
	assert.Equal(t, 1, orders[0].Quantity)
	assert.Equal(t, 80, orders[0].TotalPrice)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetOrdersByUserID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewOrderRepository(db)
	ctx := context.Background()
	userID := int64(1)

	query := `
		SELECT o\.id, o\.user_id, o\.merch_id, m\.name, o\.quantity, o\.total_price, o\.created_at
		FROM orders o
		JOIN merch m ON o\.merch_id = m\.id
		WHERE o\.user_id = \$1
		ORDER BY o\.created_at DESC`
	expectedErr := errors.New("query error")
	mock.ExpectQuery(query).WithArgs(userID).WillReturnError(expectedErr)

	orders, err := repo.GetOrdersByUserID(ctx, userID)
	assert.Error(t, err)
	assert.Nil(t, orders)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetUserByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	email := "test@example.com"

	// Подготавливаем ожидаемые строки результата.
	rows := sqlmock.NewRows([]string{"id", "username", "pass_hash", "coin_balance"}).
		AddRow(1, email, []byte("hashed-password"), 1000)
	// Ожидаем запрос с аргументом email.
	query := regexp.QuoteMeta("SELECT id, username, pass_hash, coin_balance FROM users WHERE username = $1")
	mock.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)

	user, err := repo.GetUserByEmail(ctx, email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, []byte("hashed-password"), user.PassHash)
	assert.Equal(t, 1000, user.CoinBalance)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	email := "nonexistent@example.com"

	// Эмулируем ситуацию, когда запрос возвращает 0 строк.
	rows := sqlmock.NewRows([]string{"id", "username", "pass_hash", "coin_balance"})
	query := regexp.QuoteMeta("SELECT id, username, pass_hash, coin_balance FROM users WHERE username = $1")
	mock.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)

	user, err := repo.GetUserByEmail(ctx, email)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, storage.ErrUserNotFound))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	email := "create@example.com"
	passHash := []byte("hashed")
	coinBalance := 1000

	// Подготавливаем ожидаемый запрос. Используем regexp.QuoteMeta.
	query := regexp.QuoteMeta("INSERT INTO users (username, pass_hash, coin_balance) VALUES ($1, $2, $3) RETURNING id")
	mock.ExpectQuery(query).WithArgs(email, passHash, coinBalance).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	user := &models.User{
		Email:       email,
		PassHash:    passHash,
		CoinBalance: coinBalance,
	}
	createdUser, err := repo.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), createdUser.ID)
	assert.Equal(t, email, createdUser.Email)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUserBalance_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	userID := int64(1)
	newBalance := 900

	// Ожидаем вызов Begin.
	mock.ExpectBegin()

	// Ожидаем вызов ExecContext с нужными параметрами.
	query := regexp.QuoteMeta("UPDATE users SET coin_balance = $1 WHERE id = $2")
	mock.ExpectExec(query).WithArgs(newBalance, userID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка затронута

	// Создаем транзакцию.
	tx, err := db.Begin()
	assert.NoError(t, err)

	err = repo.UpdateUserBalance(ctx, tx, userID, newBalance)
	assert.NoError(t, err)

	// Ожидаем вызов Commit.
	mock.ExpectCommit()
	err = tx.Commit()
	assert.NoError(t, err)

	// Проверяем, что все ожидания sqlmock выполнены.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestUpdateUserBalance_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	userID := int64(99)
	newBalance := 900

	// Ожидаем вызов Begin.
	mock.ExpectBegin()

	query := regexp.QuoteMeta("UPDATE users SET coin_balance = $1 WHERE id = $2")
	mock.ExpectExec(query).WithArgs(newBalance, userID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто

	tx, err := db.Begin()
	assert.NoError(t, err)

	err = repo.UpdateUserBalance(ctx, tx, userID, newBalance)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrUserNotFound))

	// Ожидаем вызов Rollback.
	mock.ExpectRollback()
	err = tx.Rollback()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByIDtx_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	userID := int64(1)
	email := "test@example.com"

	// Начинаем транзакцию.
	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)

	// Подготавливаем ожидаемые строки результата.
	rows := sqlmock.NewRows([]string{"id", "username", "pass_hash", "coin_balance"}).
		AddRow(userID, email, []byte("hashed"), 1000)
	query := regexp.QuoteMeta("SELECT id, username, pass_hash, coin_balance FROM users WHERE id = $1")
	mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)

	user, err := repo.LockUserByIDTx(ctx, tx, userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)

	mock.ExpectCommit()
	err = tx.Commit()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByIDtx_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := storage.NewUserRepository(db)
	ctx := context.Background()
	userID := int64(99)

	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "username", "pass_hash", "coin_balance"})
	query := regexp.QuoteMeta("SELECT id, username, pass_hash, coin_balance FROM users WHERE id = $1")
	mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)

	user, err := repo.LockUserByIDTx(ctx, tx, userID)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, storage.ErrUserNotFound))

	mock.ExpectRollback()
	err = tx.Rollback()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

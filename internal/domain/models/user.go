package models

// User представляет пользователя
type User struct {
	ID          int64
	Email       string
	PassHash    []byte
	CoinBalance int
}

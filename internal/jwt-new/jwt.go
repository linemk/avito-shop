package security

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/linemk/avito-shop/internal/domain/models"
)

// NewToken генерирует JWT-токен для указанного пользователя с заданным временем жизни.
func NewToken(ctx context.Context, user *models.User, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"exp":   time.Now().Add(ttl).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretStr := os.Getenv("JWT_SECRET")
	if secretStr == "" {
		return "", errors.New("JWT_SECRET environment variable is not set")
	}
	secret := []byte(secretStr)
	return token.SignedString(secret)
}

package jwtmiddleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

// NewJWTMiddleware создаёт middleware для проверки JWT, секрет берётся из переменной окружения.
func NewJWTMiddleware() func(http.Handler) http.Handler {
	// Можно также принять секрет как параметр, если не хочется брать его внутри.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is not set")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из заголовка Authorization (формат: "Bearer <token>")
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid token format", http.StatusUnauthorized)
				return
			}
			tokenStr := parts[1]

			// Парсинг и проверка токена
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				// Проверка алгоритма
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			// Извлекаем идентификатор пользователя из поля "sub"
			sub, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "invalid token claims: sub not found", http.StatusUnauthorized)
				return
			}

			userID, err := strconv.ParseInt(sub, 10, 64)
			if err != nil {
				http.Error(w, "invalid token claims: invalid user id", http.StatusUnauthorized)
				return
			}

			// Устанавливаем userID в контекст запроса
			ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext извлекает userID из контекста.
func FromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}

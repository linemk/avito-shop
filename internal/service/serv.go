package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/linemk/avito-shop/internal/domain/models"
	security "github.com/linemk/avito-shop/internal/jwt-new"
	"github.com/linemk/avito-shop/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	log      *slog.Logger
	userRepo storage.UserStorage
	tokenTTL time.Duration
}

func NewAuthService(log *slog.Logger, userRepo storage.UserStorage, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		log:      log,
		userRepo: userRepo,
		tokenTTL: tokenTTL,
	}
}

type AuthServiceInterface interface {
	Login(ctx context.Context, username, password string) (string, error)
}

// Login осуществляет аутентификацию пользователя.
// Если пользователь не найден, он создаётся (при этом пароль хэшируется через bcrypt, который автоматически добавляет соль).
// Если пользователь найден, введённый пароль сравнивается с сохранённым хэшированным значением.
// После успешной проверки генерируется JWT-токен (секрет для подписи берется из переменной окружения).
func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "auth.Login"
	logger := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	logger.Info("checking user")

	// Попытка получить пользователя по email из базы
	user, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			logger.Info("user not found, creating new user")
			// Хеширование пароля с помощью bcrypt (автоматически добавляет соль)
			passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				logger.Error("failed to hash password", slog.Any("error", err))
				return "", fmt.Errorf("%s: failed to hash password: %w", op, err)
			}
			newUser := &models.User{
				Email:       email,
				PassHash:    passHash,
				CoinBalance: 1000, // начальный баланс
			}
			user, err = a.userRepo.CreateUser(ctx, newUser)
			if err != nil {
				logger.Error("failed to create user", slog.Any("error", err))
				return "", fmt.Errorf("%s: failed to create user: %w", op, err)
			}
		} else {
			logger.Error("failed to get user", slog.Any("error", err))
			return "", fmt.Errorf("%s: failed to get user: %w", op, err)
		}
	} else {
		// Если пользователь найден, сравниваем введённый пароль с хэшированным паролем
		if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
			logger.Warn("invalid password")
			return "", fmt.Errorf("%s: invalid credentials: %w", op, err)
		}
	}

	// Генерация JWT-токена. Функция auth.NewToken внутри сама загружает секрет из переменной окружения JWT_SECRET.
	token, err := security.NewToken(ctx, user, a.tokenTTL)
	if err != nil {
		logger.Error("failed to generate token", slog.Any("error", err))
		return "", fmt.Errorf("%s: failed to generate token: %w", op, err)
	}

	logger.Info("user logged in successfully", slog.Int64("userID", user.ID))
	return token, nil
}

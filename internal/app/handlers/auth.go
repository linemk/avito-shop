package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/linemk/avito-shop/internal/service"
)

// AuthRequest представляет структуру запроса для аутентификации с тегами валидации
type AuthRequest struct {
	Username string `json:"username" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// AuthResponse представляет структуру ответа с JWT-токеном
type AuthResponse struct {
	Token string `json:"token"`
}

var validate = validator.New()

// AuthHandler – HTTP-обработчик для аутентификации, принимает логгер и экземпляр AuthService
func AuthHandler(log *slog.Logger, authService service.AuthServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.AuthHandler"
		logger := log.With(slog.String("op", op))

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("invalid request: decoding error", slog.Any("error", err))
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Валидация структуры запроса с использованием validator
		if err := validate.Struct(req); err != nil {
			logger.Error("invalid request: validation error", slog.Any("error", err))
			http.Error(w, "validation error", http.StatusBadRequest)
			return
		}

		// Вызов бизнес-логики для аутентификации
		token, err := authService.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			logger.Error("login failed", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Формирование и отправка ответа с JWT-токеном
		resp := AuthResponse{Token: token}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode response", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

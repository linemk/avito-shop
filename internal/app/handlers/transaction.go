package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/linemk/avito-shop/internal/jwt-new/jwtmiddleware"
	"github.com/linemk/avito-shop/internal/service"
)

// SendCoinRequest представляет входной JSON для перевода монет.
type SendCoinRequest struct {
	ToUser string `json:"toUser" validate:"required,email"`
	Amount int    `json:"amount" validate:"required,gt=0"`
}

// SendCoinResponse представляет ответ при успешном переводе.
type SendCoinResponse struct {
	Message string `json:"message"`
}

// SendCoinHandler обрабатывает запрос POST /api/sendCoin.
func SendCoinHandler(log *slog.Logger, sendCoinService service.SendCoinService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.SendCoinHandler"
		logger := log.With(slog.String("op", op))

		var req SendCoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("invalid request: decoding error", slog.Any("error", err))
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Здесь можно добавить валидацию через validator, если необходимо

		// Извлекаем userID отправителя из контекста (установленного JWT middleware)
		userID, ok := jwtmiddleware.FromContext(r.Context())
		if !ok {
			logger.Error("userID not found in context")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Вызываем бизнес-логику для перевода монет
		if err := sendCoinService.SendCoin(r.Context(), userID, req.ToUser, req.Amount); err != nil {
			logger.Error("failed to send coin", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := SendCoinResponse{Message: "Coins transferred successfully"}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode response", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

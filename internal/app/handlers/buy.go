package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linemk/avito-shop/internal/jwt-new/jwtmiddleware"
	"github.com/linemk/avito-shop/internal/service"
)

// BuyResponse — структура ответа при успешной покупке.
type BuyResponse struct {
	Message string `json:"message"`
}

// BuyHandler обрабатывает запрос GET /api/buy/{item}
func BuyHandler(log *slog.Logger, buyService service.BuyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.BuyHandler"
		logger := log.With(slog.String("op", op))

		// Извлекаем название товара из URL
		item := chi.URLParam(r, "item")
		if item == "" {
			logger.Error("item parameter is missing")
			http.Error(w, "item parameter is required", http.StatusBadRequest)
			return
		}

		// Извлекаем userID из контекста (установленный JWT middleware)
		userID, ok := jwtmiddleware.FromContext(r.Context())
		if !ok {
			logger.Error("userID not found in context")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Вызываем бизнес-логику для покупки
		if err := buyService.Buy(r.Context(), userID, item); err != nil {
			logger.Error("failed to complete purchase", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Формируем ответ
		resp := BuyResponse{Message: "Item purchased successfully"}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode response", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

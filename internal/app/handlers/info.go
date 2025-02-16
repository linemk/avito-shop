package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/linemk/avito-shop/internal/jwt-new/jwtmiddleware"
	"github.com/linemk/avito-shop/internal/service"
)

// InfoResponse — структура ответа, соответствующая OpenAPI.
type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []HistoryEntry `json:"received"`
	Sent     []HistoryEntry `json:"sent"`
}

type HistoryEntry struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}

// InfoHandler обрабатывает запрос GET /api/info.
// Он извлекает идентификатор пользователя из контекста (установленный JWT‑middleware),
// затем вызывает сервис InfoService для получения информации о балансе, инвентаре и истории транзакций
func InfoHandler(log *slog.Logger, infoService service.InfoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.InfoHandler"
		logger := log.With(slog.String("op", op))

		// Извлечение userID из контекста, который установил JWT‑middleware
		userID, ok := jwtmiddleware.FromContext(r.Context())
		if !ok {
			logger.Error("userID not found in context")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Вызов бизнес‑логики для получения информации о пользователе
		info, err := infoService.GetInfo(r.Context(), userID)
		if err != nil {
			logger.Error("failed to get info", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Отправка JSON‑ответа клиенту
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(info); err != nil {
			logger.Error("failed to encode response", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

package urllog

import (
	"log/slog"
	"net/http"
)

func CustomLoggerMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("request received",
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
			)
			next.ServeHTTP(w, r)
		})
	}
}

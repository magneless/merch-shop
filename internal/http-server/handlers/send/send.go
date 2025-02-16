package send

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.buy.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		_ = log
	}
}
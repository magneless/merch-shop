package buy

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	mwAuth "github.com/magneless/merch-shop/internal/http-server/middleware/auth"
	"github.com/magneless/merch-shop/internal/lib/api/response"
	"github.com/magneless/merch-shop/internal/lib/logger/sl"
)

type MerchPurchaser interface {
	PurchaseMerch(username, merchName string, quantity int) error
}

// добавить специализированный ответ, если недостаточно средств
func New(log *slog.Logger, merchPurchaser MerchPurchaser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.buy.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		item := chi.URLParam(r, "item")
		if item == "" {
			log.Error("item is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse{Error: "chose item"})
		}

		username, ok := r.Context().Value(mwAuth.UsernameKey).(string)
		if !ok {
			log.Error("failed to get username form context")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
			return
		}

		err := merchPurchaser.PurchaseMerch(username, item, 1)
		if err != nil {
			log.Error("failed to purchase merch from db", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
			return
		}

		render.Status(r, http.StatusOK)

		log.Info("user bought item", slog.String("username", username))
	}
}

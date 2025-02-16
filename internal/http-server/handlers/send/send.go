package send

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	mwAuth "github.com/magneless/merch-shop/internal/http-server/middleware/auth"
	"github.com/magneless/merch-shop/internal/lib/api/response"
	"github.com/magneless/merch-shop/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}
type CoinsSender interface {
	SendCoins(senderUsername, receiverUsername string, amount int) error
}

func New(log *slog.Logger, coinsSender CoinsSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.buy.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		username, ok := r.Context().Value(mwAuth.UsernameKey).(string)
		if !ok {
			log.Error("failed to get username form context")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
			return
		}

		var req SendCoinRequest

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse{Error: "empty request"})
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "iternal error"})
			return
		}

		log.Info("request body decoded ", slog.Any("request", slog.String("username", username)))

		err = coinsSender.SendCoins(username, req.ToUser, req.Amount)
		if err != nil {
			log.Error("failed to send coins", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
		}

		log.Info("Coins successfuly sent", slog.Any("particapants", map[string]string{
			"sender": username,
			"receiver": req.ToUser,
		}))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.MessageResponse{Message: "coins successfully sent"})
	}
}

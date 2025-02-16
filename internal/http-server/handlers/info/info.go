package info

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	mwAuth "github.com/magneless/merch-shop/internal/http-server/middleware/auth"
	"github.com/magneless/merch-shop/internal/lib/api/response"
	"github.com/magneless/merch-shop/internal/lib/logger/sl"
	"github.com/magneless/merch-shop/internal/models"
)

type InfoGetter interface {
	GetInventory(userID int) ([]models.InventoryItem, error)
	GetReceivedTransactions(userID int, toUsername string) ([]models.CoinTransaction, error)
	GetSentTransactions(userID int, fromUsername string) ([]models.CoinTransaction, error)
	GetBalanceAndId(username string) (int, int, error)
}

func New(log *slog.Logger, infoGetter InfoGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.info.New"

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

		userID, balance, err := infoGetter.GetBalanceAndId(username)
		if err != nil {
			log.Error("failed to get balance or id from bd", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
		}

		sent, err := infoGetter.GetSentTransactions(userID, username)
		if err != nil {
			log.Error("failed to get sent transactions from bd", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
		}

		received, err := infoGetter.GetReceivedTransactions(userID, username)
		if err != nil {
			log.Error("failed to get received transactions from bd", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
		}

		inventory, err := infoGetter.GetInventory(userID)
		if err != nil {
			log.Error("failed to get inventory from bd", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
		}

		log.Info("user got his info", slog.String("username", username))

		render.JSON(w, r, response.InfoResponse{
			Coins:     balance,
			Inventory: inventory,
			CoinHistory: models.CoinHistory{
				Sent:     sent,
				Received: received,
			},
		})
	}
}

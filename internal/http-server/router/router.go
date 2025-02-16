package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	auth "github.com/magneless/merch-shop/internal/http-server/handlers/auth"
	"github.com/magneless/merch-shop/internal/http-server/handlers/buy"
	"github.com/magneless/merch-shop/internal/http-server/handlers/info"
	"github.com/magneless/merch-shop/internal/http-server/handlers/send"
	mwAuth "github.com/magneless/merch-shop/internal/http-server/middleware/auth"
	mwLogger "github.com/magneless/merch-shop/internal/http-server/middleware/logger"
	"github.com/magneless/merch-shop/internal/models"
)

type Auth interface {
	GetUser(username, passwordHash string) error
}

type Info interface {
	GetInventory(userID int) ([]models.InventoryItem, error)
	GetReceivedTransactions(userID int, toUsername string) ([]models.CoinTransaction, error)
	GetSentTransactions(userID int, fromUsername string) ([]models.CoinTransaction, error)
	GetBalanceAndId(username string) (int, int, error)
}

type Buy interface {
	PurchaseMerch(username, merchName string, quantity int) error
}

type Send interface {
	SendCoins(senderUsername, receiverUsername string, amount int) error	
}

type Repository interface {
	Auth
	Info
	Buy
	Send
}

func New(log *slog.Logger, repo Repository) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(log))
	r.Use(middleware.Recoverer)

	r.Post("/api/auth", auth.New(log, repo))

	r.Route("/api", func(r chi.Router) {
		r.Use(mwAuth.New(log))

		r.Get("/info", info.New(log, repo))
		r.Post("/sendCoin", send.New(log, repo))
		r.Get("/buy/{item}", buy.New(log, repo))
	})

	return r
}

package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	auth "github.com/magneless/merch-shop/internal/http-server/handlers/auth"
	"github.com/magneless/merch-shop/internal/http-server/handlers/buy"
	"github.com/magneless/merch-shop/internal/http-server/handlers/info"
	"github.com/magneless/merch-shop/internal/http-server/handlers/send"
	mwLogger "github.com/magneless/merch-shop/internal/http-server/middleware/logger"
	mwAuth "github.com/magneless/merch-shop/internal/http-server/middleware/auth"
)

type Auth interface {
	GetUser(username, passwordHash string) error
}

type Repository interface {
	Auth
}

func New(log *slog.Logger, repo Repository) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(log))
	r.Use(middleware.Recoverer)

	r.Post("/api/auth", auth.New(log, repo))

	r.Route("/api", func(r chi.Router) {
		r.Use(mwAuth.New(log))

		r.Get("/info", info.New(log))
		r.Post("/sendCoin", send.New(log))
		r.Get("/buy/{item}", buy.New(log))
	})

	return r
}

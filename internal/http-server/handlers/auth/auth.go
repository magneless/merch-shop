package auth

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/magneless/merch-shop/internal/lib/api/response"
	"github.com/magneless/merch-shop/internal/lib/hashing"
	jwt_token "github.com/magneless/merch-shop/internal/lib/jwt"
	"github.com/magneless/merch-shop/internal/lib/logger/sl"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserGetter interface {
	GetUser(username, passwordHash string) error
}

func New(log *slog.Logger, userGetter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req AuthRequest

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse{Error: "request body is empty"})
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
			return
		}

		log.Info("request body decoded", slog.String("username", req.Username))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		passwordHash, err := hashing.HashPassword(req.Password)
		if err != nil {
			log.Error("failed to hash password", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
			return
		}

		// по api я совместил регистрацию и вход в аккаунт
		err = userGetter.GetUser(req.Username, passwordHash)
		if err != nil {
			log.Error("failed to auth user", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse{Error: "wrong password or login"})
			return
		}

		accessToken, err := jwt_token.GenerateAccessToken(req.Username)
		if err != nil {
			log.Error("failed to generate access token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse{Error: "internal error"})
			return
		}

		log.Info("user authed", slog.String("username", req.Username))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.AuthResponse{Token: accessToken})
	}
}

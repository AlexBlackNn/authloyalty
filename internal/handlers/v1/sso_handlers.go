package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/services/auth_service"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type AuthHandlers struct {
	log         *slog.Logger
	authService *auth_service.Auth
	auth        auth_service.AuthorizationInterface
}

func New(log *slog.Logger, authService *auth_service.Auth) AuthHandlers {
	return AuthHandlers{log: log, authService: authService}
}

func (a *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responseError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var reqLogin Login
	err := render.DecodeJSON(r.Body, &reqLogin)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Post with empty body
			responseError(w, r, http.StatusBadRequest, "empty request")
			return
		}
		responseError(w, r, http.StatusBadRequest, "failed to decode request")
		return
	}
	if err = validator.New().Struct(reqLogin); err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			errorText := ValidationError(validateErr)
			responseError(w, r, http.StatusBadRequest, errorText)
			return
		}
		responseError(w, r, http.StatusUnprocessableEntity, "failed to validate request")
		return
	}

	ctx, cancel := context.WithTimeoutCause(r.Context(), 300*time.Millisecond, errors.New("updateMetric timeout"))
	defer cancel()

	accessToken, refreshToken, err := a.auth.Login(
		ctx, reqLogin.Email, reqLogin.Password,
	)
	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, auth_service.ErrInvalidCredentials) {
			responseError(w, r, http.StatusNotFound, err.Error())
			return
		}
		responseError(w, r, http.StatusInternalServerError, "internal server error")
		return
	}
	fmt.Println(accessToken, refreshToken)
	responseOK(w, r)
}

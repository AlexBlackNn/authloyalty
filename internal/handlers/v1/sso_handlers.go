package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/services/auth_service"
	"github.com/AlexBlackNn/authloyalty/pkg/storage"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type AuthHandlers struct {
	log  *slog.Logger
	auth *auth_service.Auth
}

func New(log *slog.Logger, authService *auth_service.Auth) AuthHandlers {
	return AuthHandlers{log: log, auth: authService}
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

	fmt.Println("111111111111111111111111111", reqLogin, reqLogin.Email, reqLogin.Password)

	accessToken, refreshToken, err := a.auth.Login(
		ctx, reqLogin.Email, reqLogin.Password,
	)
	fmt.Println("222222222222222222222222222222", reqLogin, reqLogin.Email, reqLogin.Password)

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

func (a *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responseError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var reqRegister Register
	err := render.DecodeJSON(r.Body, &reqRegister)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Post with empty body
			responseError(w, r, http.StatusBadRequest, "empty request")
			return
		}
		responseError(w, r, http.StatusBadRequest, "failed to decode request")
		return
	}
	if err = validator.New().Struct(reqRegister); err != nil {
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

	fmt.Println("111111111111111111111111111", reqRegister, reqRegister.Email, reqRegister.Password)

	userID, err := a.auth.Register(
		ctx, reqRegister.Email, reqRegister.Password,
	)

	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, storage.ErrUserExists) {
			responseError(w, r, http.StatusConflict, err.Error())
			return
		}
		responseError(w, r, http.StatusInternalServerError, "internal server error")
		return
	}
	fmt.Println(userID)
	responseOK(w, r)
}

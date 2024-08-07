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

// @Summary Login
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body Login true "Login request"
// @Success 200 {object} Response "Login successful"
// @Failure 400 {object} Response "Bad request"
// @Failure 401 {object} Response "Unauthorized"
// @Failure 404 {object} Response "User not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/login [post]
func (a *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responseError(
			w, r, http.StatusMethodNotAllowed, "method not allowed",
		)
		return
	}

	var reqLogin Login
	err := render.DecodeJSON(r.Body, &reqLogin)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Post with empty body
			responseError(
				w, r, http.StatusBadRequest, "empty request",
			)
			return
		}
		responseError(
			w, r, http.StatusBadRequest, "failed to decode request",
		)
		return
	}
	if err = validator.New().Struct(reqLogin); err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			errorText := ValidationError(validateErr)
			responseError(
				w, r, http.StatusBadRequest, errorText,
			)
			return
		}
		responseError(
			w, r, http.StatusUnprocessableEntity, "failed to validate request",
		)
		return
	}

	ctx, cancel := context.WithTimeoutCause(
		r.Context(), 300*time.Millisecond, errors.New("updateMetric timeout"),
	)
	defer cancel()

	accessToken, refreshToken, err := a.auth.Login(
		ctx, reqLogin.Email, reqLogin.Password,
	)

	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, auth_service.ErrInvalidCredentials) {
			responseError(
				w, r, http.StatusNotFound, err.Error(),
			)
			return
		}
		responseError(
			w, r, http.StatusInternalServerError, "internal server error",
		)
		return
	}
	responseAccessRefresh(
		w, r, http.StatusOK, "Ok", accessToken, refreshToken,
	)
}

// @Summary Logout
// @Description Logout from current session. Frontend needs to send access and then refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body Logout true "Logout request"
// @Success 200 {object} Response "Logout successful"
// @Failure 400 {object} Response "Bad request"
// @Failure 401 {object} Response "Unauthorized"
// @Failure 404 {object} Response "User not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/logout [post]
func (a *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responseError(
			w, r, http.StatusMethodNotAllowed, "method not allowed",
		)
		return
	}

	var reqLogout Logout
	err := render.DecodeJSON(r.Body, &reqLogout)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Post with empty body
			responseError(w, r, http.StatusBadRequest, "empty request")
			return
		}
		responseError(w, r, http.StatusBadRequest, "failed to decode request")
		return
	}
	if err = validator.New().Struct(reqLogout); err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			errorText := ValidationError(validateErr)
			responseError(
				w, r, http.StatusBadRequest, errorText,
			)
			return
		}
		responseError(
			w, r, http.StatusUnprocessableEntity, "failed to validate request",
		)
		return
	}

	ctx, cancel := context.WithTimeoutCause(
		r.Context(), 300*time.Millisecond, errors.New("updateMetric timeout"),
	)
	defer cancel()

	_, err = a.auth.Logout(ctx, reqLogout.Token)
	if err != nil {
		responseError(
			w, r, http.StatusInternalServerError, "internal server error",
		)
		return
	}
	responseOK(w, r)
}

// @Summary Registration
// @Description User registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body Register Register "Register request"
// @Success 200 {object} Response "Register successful"
// @Failure 400 {object} Response "Bad request"
// @Failure 401 {object} Response "Unauthorized"
// @Failure 404 {object} Response "User not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /auth/registration [post]
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
			responseError(
				w, r, http.StatusBadRequest, "empty request",
			)
			return
		}
		responseError(
			w, r, http.StatusBadRequest, "failed to decode request",
		)
		return
	}
	if err = validator.New().Struct(reqRegister); err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			errorText := ValidationError(validateErr)
			responseError(
				w, r, http.StatusBadRequest, errorText,
			)
			return
		}
		responseError(
			w, r, http.StatusUnprocessableEntity, "failed to validate request",
		)
		return
	}

	ctx, cancel := context.WithTimeoutCause(
		r.Context(), 300*time.Millisecond, errors.New("updateMetric timeout"),
	)
	defer cancel()

	_, err = a.auth.Register(
		ctx, reqRegister.Email, reqRegister.Password,
	)

	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, storage.ErrUserExists) {
			responseError(
				w, r, http.StatusConflict, err.Error(),
			)
			return
		}
		responseError(
			w, r, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	accessToken, refreshToken, err := a.auth.Login(
		ctx, reqRegister.Email, reqRegister.Password,
	)

	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, auth_service.ErrInvalidCredentials) {
			responseError(
				w, r, http.StatusNotFound, err.Error(),
			)
			return
		}
		responseError(
			w, r, http.StatusInternalServerError, "internal server error",
		)
		return
	}
	responseAccessRefresh(
		w, r, http.StatusOK, "Ok", accessToken, refreshToken,
	)
}

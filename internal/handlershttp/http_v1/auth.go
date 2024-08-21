package http_v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	"github.com/AlexBlackNn/authloyalty/pkg/storage"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type AuthHandlers struct {
	log  *slog.Logger
	auth *authservice.Auth
	cfg  *config.Config
}

func New(log *slog.Logger, cfg *config.Config, authService *authservice.Auth) AuthHandlers {
	return AuthHandlers{log: log, cfg: cfg, auth: authService}
}

// EasyJSONUnmarshaler provides ability easyjson lib to work with generic type.
// In case of using "err := render.DecodeJSON(r.Body, &reqData)" it can be deleted.
type EasyJSONUnmarshaler interface {
	UnmarshalJSON(data []byte) error
}

// ValidateRequest validates post body. In case of using "err := render.DecodeJSON(r.Body, &reqData)"
// can be written as ValidateRequest[T Login | Logout | Refresh | Register](...). In case of using easyjson Login,
// Logout, Refresh, Register have  UnmarshalJSON method after code generation. EasyJSONUnmarshaler must be use here to
// work with easyjson.
func ValidateRequest[T EasyJSONUnmarshaler](w http.ResponseWriter, r *http.Request, reqData T) (T, error) {
	if r.Method != http.MethodPost {
		models.ResponseErrorNowAllowed(w, "only POST method allowed")
		return reqData, errors.New("method not allowed")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.ResponseErrorBadRequest(w, "failed to read body")
		return reqData, errors.New("failed to read body")
	}
	err = reqData.UnmarshalJSON(body)
	if err != nil {
		models.ResponseErrorBadRequest(w, "failed to decode request")
		return reqData, errors.New("failed to decode request")
	}
	if err = validator.New().Struct(reqData); err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			models.ResponseErrorBadRequest(w, models.ValidationError(validateErr))
			return reqData, errors.New("validation error")
		}
		models.ResponseErrorBadRequest(w, "bad request")
		return reqData, errors.New("bad request")
	}
	return reqData, nil
}

// @Summary Login
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.Login true "Login request"
// @Success 201 {object} models.Response "Login successful"
// @Router /auth/login [post]
func (a *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {

	reqData, err := ValidateRequest[*models.Login](w, r, &models.Login{})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeoutCause(
		r.Context(),
		time.Duration(a.cfg.ServerHandlersTimeouts.LoginTimeoutMs)*time.Millisecond,
		errors.New("updateMetric timeout"),
	)
	defer cancel()

	accessToken, refreshToken, err := a.auth.Login(ctx, reqData)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			models.ResponseErrorNotFound(w, "user not found")
			return
		}
		models.ResponseErrorInternal(w, "internal server error")
		return
	}
	models.ResponseOKAccessRefresh(w, accessToken, refreshToken)
}

// @Summary Logout
// @Description Logout from current session. Frontend needs to send access and then refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.Logout true "Logout request"
// @Success 200 {object} models.Response "Logout successful"
// @Router /auth/logout [post]
func (a *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	reqData, err := ValidateRequest[*models.Logout](w, r, &models.Logout{})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeoutCause(
		r.Context(),
		time.Duration(a.cfg.ServerHandlersTimeouts.LogoutTimeoutMs)*time.Millisecond,
		errors.New("updateMetric timeout"),
	)
	defer cancel()

	_, err = a.auth.Logout(ctx, reqData)
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrUserNotFound):
			models.ResponseErrorNotFound(w, "user not found")
		case errors.Is(err, authservice.ErrTokenRevoked):
			models.ResponseErrorStatusConflict(w, "token revoked")
		case errors.Is(err, authservice.ErrTokenParsing):
			models.ResponseErrorBadRequest(w, "token error")
		case errors.Is(err, authservice.ErrTokenTTLExpired):
			models.ResponseErrorStatusConflict(w, "token ttl expired")
		default:
			models.ResponseErrorInternal(w, "internal server error")
		}
		return
	}
	models.ResponseOK(w)
}

// @Summary Registration
// @Description User registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.Register true "Register request"
// @Success 201 {object} models.Response "Register successful"
// @Router /auth/registration [post]
func (a *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	reqData, err := ValidateRequest[*models.Register](w, r, &models.Register{})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeoutCause(
		r.Context(),
		time.Duration(a.cfg.ServerHandlersTimeouts.RegisterTimeoutMs)*time.Millisecond,
		errors.New("updateMetric timeout"),
	)
	defer cancel()

	ctx, _, err = a.auth.Register(ctx, reqData)
	if err != nil {
		fmt.Println(err.Error())
		// TODO change to service error
		if errors.Is(err, storage.ErrUserExists) {
			models.ResponseErrorStatusConflict(w, "user already exists")
			return
		}
		models.ResponseErrorInternal(w, "internal server error")
		return
	}

	accessToken, refreshToken, err := a.auth.Login(
		ctx, &models.Login{Email: reqData.Email, Password: reqData.Password},
	)

	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			models.ResponseErrorNotFound(w, "user not found")
			return
		}
		models.ResponseErrorInternal(w, "internal server error")
		return
	}
	models.ResponseOKAccessRefresh(w, accessToken, refreshToken)
}

// @Summary Refresh
// @Description
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.Refresh true "Refresh request"
// @Success 201 {object} models.Response "Refresh successful"
// @Router /auth/refresh [post]
func (a *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	reqData, err := ValidateRequest[*models.Refresh](w, r, &models.Refresh{})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeoutCause(
		r.Context(),
		time.Duration(a.cfg.ServerHandlersTimeouts.RefreshTimeoutMs)*time.Millisecond,
		errors.New("updateMetric timeout"),
	)
	defer cancel()

	accessToken, refreshToken, err := a.auth.Refresh(ctx, reqData)

	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrUserNotFound):
			models.ResponseErrorNotFound(w, "user not found")
		case errors.Is(err, authservice.ErrTokenWrongType):
			models.ResponseErrorStatusConflict(w, "token wrong type, expected refresh")
		case errors.Is(err, authservice.ErrTokenRevoked):
			models.ResponseErrorStatusConflict(w, "token revoked")
		case errors.Is(err, authservice.ErrTokenParsing):
			models.ResponseErrorBadRequest(w, "token error")
		case errors.Is(err, authservice.ErrTokenTTLExpired):
			models.ResponseErrorStatusConflict(w, "token ttl expired")
		default:
			models.ResponseErrorInternal(w, "internal server error")
		}
		return
	}
	models.ResponseOKAccessRefresh(w, accessToken, refreshToken)
}

package v1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
)

type loyaltyService interface {
	AddLoyalty(
		ctx context.Context,
		reqData *domain.UserLoyalty,
	) (domain.UserLoyalty, error)
	SubLoyalty(
		ctx context.Context,
		reqData *domain.UserLoyalty,
	) (domain.UserLoyalty, error)
	GetLoyalty(
		ctx context.Context,
		reqData *domain.UserLoyalty,
	) (domain.UserLoyalty, error)
}

type LoyaltyHandlers struct {
	log     *slog.Logger
	loyalty loyaltyService
	cfg     *config.Config
}

func New(log *slog.Logger, cfg *config.Config, loyalty loyaltyService) LoyaltyHandlers {
	return LoyaltyHandlers{log: log, cfg: cfg, loyalty: loyalty}
}

func ctxWithTimeoutCause(r *http.Request, cfg *config.Config, textError string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeoutCause(
		r.Context(),
		time.Duration(cfg.ServerHandlersTimeouts.LoginTimeoutMs)*time.Millisecond,
		errors.New(textError),
	)
	return ctx, cancel
}

// @Summary AddLoyalty
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags Loyalty
// @Accept json
// @Produce json
// @Param body body models.UserLoyalty true "UserLoyalty request"
// @Success 201 {object} models.Response "Add loyalty successful"
// @Router /loyalty [post]
// @Security BearerAuth
func (l *LoyaltyHandlers) AddLoyalty(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := ctxWithTimeoutCause(r, l.cfg, "login timeout")
	defer cancel()

	if r.Method != http.MethodPost {
		dto.ResponseErrorNowAllowed(w, "only POST method allowed")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		dto.ResponseErrorBadRequest(w, "failed to read body")
	}
	reqData := &dto.UserLoyalty{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err = json.Unmarshal(body, reqData)
	if err != nil {
		fmt.Println(reqData)
		dto.ResponseErrorBadRequest(w, "failed to decode body")
		return
	}

	if err = validator.New().Struct(reqData); err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			dto.ResponseErrorBadRequest(w, dto.ValidationError(validateErr))
		}
		dto.ResponseErrorBadRequest(w, "bad request")
	}

	loyalty, err := l.loyalty.AddLoyalty(ctx, &domain.UserLoyalty{UUID: reqData.UUID, Value: reqData.Value})
	if err != nil {
		dto.ResponseErrorInternal(w, "internal server error")
		return
	}
	dto.ResponseOKLoyalty(w, loyalty.UUID, loyalty.Value)
}

// @Summary AddLoyalty
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags Loyalty
// @Accept json
// @Produce json
// @Param body body models.UserLoyalty true "UserLoyalty request"
// @Success 200 {object} models.Response "Get loyalty successful"
// @Router /loyalty [get]
// @Security BearerAuth
func (l *LoyaltyHandlers) GetLoyalty(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := ctxWithTimeoutCause(r, l.cfg, "login timeout")
	defer cancel()

	if r.Method != http.MethodGet {
		dto.ResponseErrorNowAllowed(w, "only POST method allowed")
	}

	currentUuid := chi.URLParam(r, "uuid")

	loyalty, err := l.loyalty.GetLoyalty(ctx, &domain.UserLoyalty{UUID: currentUuid, Value: 0})
	if err != nil {
		dto.ResponseErrorInternal(w, "internal server error")
		return
	}
	dto.ResponseOKLoyalty(w, loyalty.UUID, loyalty.Value)
}

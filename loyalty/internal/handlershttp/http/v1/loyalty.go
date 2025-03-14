package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/dto"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/jwt"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/services/loyaltyservice"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/ssoclient"
	"go.opentelemetry.io/otel"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
)

type loyaltyService interface {
	AddLoyalty(
		ctx context.Context,
		reqData *domain.UserLoyalty,
	) (*domain.UserLoyalty, error)
	GetLoyalty(
		ctx context.Context,
		reqData *domain.UserLoyalty,
	) (*domain.UserLoyalty, error)
}

type LoyaltyHandlers struct {
	log       *slog.Logger
	cfg       *config.Config
	loyalty   loyaltyService
	ssoClient *ssoclient.SSOClient
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	loyalty loyaltyService,
) LoyaltyHandlers {
	ssoClient, err := ssoclient.New(cfg)
	if err != nil {
		log.Error("can't create SSO client")
	}
	return LoyaltyHandlers{
		log:       log,
		cfg:       cfg,
		ssoClient: ssoClient,
		loyalty:   loyalty,
	}
}

func ctxWithTimeoutCause(
	r *http.Request,
	cfg *config.Config,
	textError string,
) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeoutCause(
		r.Context(),
		time.Duration(cfg.ServerHandlersTimeouts.LoginTimeoutMs)*time.Millisecond,
		errors.New(textError),
	)
	return ctx, cancel
}

var tracer = otel.Tracer("loyalty service")

// @Summary AddLoyalty
// @Description Add Loyalty
// @Tags Loyalty
// @Accept json
// @Produce json
// @Param body body dto.UserLoyalty true "UserLoyalty request"
// @Success 201 {object} dto.Response "Add loyalty successful"
// @Router /loyalty [post]
// @Security BearerAuth
func (l *LoyaltyHandlers) AddLoyalty(w http.ResponseWriter, r *http.Request) {
	reqData, err := handleAddLoyaltyBadRequest(w, r, &dto.UserLoyalty{})
	if err != nil {
		return
	}

	ctx, cancel := ctxWithTimeoutCause(r, l.cfg, "add loyalty")
	defer cancel()

	tokenString := r.Header.Get("Authorization")
	token := strings.TrimPrefix(tokenString, "Bearer")
	token = strings.TrimPrefix(token, " ")
	if !l.ssoClient.IsJWTValid(ctx, tracer, token) {
		dto.ResponseErrorBadRequest(w, "jwt token invalid")
		return
	}
	uuid, _, err := jwt.Parse(token)
	if err != nil {
		dto.ResponseErrorBadRequest(w, "jwt token parsing failed")
		return
	}

	var userLoyalty *domain.UserLoyalty

	isAdmin := l.ssoClient.IsAdmin(ctx, tracer, uuid)
	// only admins can deposit and withdraw loyalty using uuid in post request
	if isAdmin {
		userLoyalty = &domain.UserLoyalty{
			UUID:      reqData.UUID,
			Operation: reqData.Operation,
			Comment:   reqData.Comment,
			Balance:   reqData.Balance,
		}
	} else if reqData.Operation == "d" {
		dto.ResponseErrorBadRequest(w, "only admins can deposit loyalty")
		return
	} else {
		// users can only withdraw loyalty from their own account (uuid extracted from jwt)
		userLoyalty = &domain.UserLoyalty{
			UUID:      uuid,
			Operation: reqData.Operation,
			Comment:   reqData.Comment,
			Balance:   reqData.Balance,
		}
	}

	loyalty, err := l.loyalty.AddLoyalty(ctx, userLoyalty)

	if err != nil {
		if errors.Is(err, loyaltyservice.ErrNegativeBalance) {
			dto.ResponseErrorBadRequest(w, "withdraw such amount of loyalty leads to negative balance")
			return
		}
		if errors.Is(err, loyaltyservice.ErrUserNotFound) {
			dto.ResponseErrorBadRequest(w, "user not found")
			return
		}
		dto.ResponseErrorInternal(w, "internal server error")
		return
	}
	dto.ResponseOKLoyalty(w, loyalty.UUID, loyalty.Balance)
}

// @Summary GetLoyalty
// @Description Get Loyalty
// @Tags Loyalty
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response "Get loyalty successful"
// @Router /loyalty/{uuid} [get]
// @Param uuid path string true "User UUID"
// @Security BearerAuth
func (l *LoyaltyHandlers) GetLoyalty(w http.ResponseWriter, r *http.Request) {
	userLoyalty, err := handleGetLoyaltyBadRequest(w, r)
	ctx, cancel := ctxWithTimeoutCause(r, l.cfg, "login timeout")
	defer cancel()
	loyalty, err := l.loyalty.GetLoyalty(ctx, userLoyalty)

	if err != nil {
		if errors.Is(err, loyaltyservice.ErrUserNotFound) {
			dto.ResponseErrorNotFound(w, "user not found")
			return
		}
		dto.ResponseErrorInternal(w, "internal server error")
		return
	}
	dto.ResponseOKLoyalty(w, loyalty.UUID, loyalty.Balance)
}

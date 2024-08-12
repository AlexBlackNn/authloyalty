package http_v1

import (
	"context"
	"errors"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	"log/slog"
	"net/http"
	"time"
)

type HealthHandlers struct {
	log         *slog.Logger
	authservice *authservice.Auth
}

func NewHealth(log *slog.Logger, authservice *authservice.Auth) HealthHandlers {
	return HealthHandlers{log: log, authservice: authservice}
}

type Request struct {
	Expression string `json:"expression" validate:"required"`
}

// @Summary Проверка готовности приложения
// @Description Определяет можно ли подавать трафик на сервис
// @Tags Health
// @Produce json
// @Success 200 {object} models.Response
// @Router /auth/ready [get]
func (m *HealthHandlers) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		models.ResponseErrorNowAllowed(w, "only GET method allowed")
		return
	}
	ctx, cancel := context.WithTimeoutCause(r.Context(), 300*time.Millisecond, errors.New("readinessProbe timeout"))
	defer cancel()

	ctx, err := m.authservice.HealthCheck(ctx)

	if err != nil {
		models.ResponseErrorInternal(w, "internal server error")
		return
	}
	models.ResponseOK(w)
}

// @Summary Проверка, что приложение живо
// @Description Определяет, нужно ли перезагрузить сервис
// @Tags Health
// @Produce json
// @Success 200 {object} models.Response
// @Router /auth/healthz [get]
func (m *HealthHandlers) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		models.ResponseErrorNowAllowed(w, "only GET method allowed")
		return
	}
	models.ResponseOK(w)
}

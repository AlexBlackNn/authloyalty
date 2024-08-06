package v1

import (
	"context"
	"errors"
	authservice "github.com/AlexBlackNn/authloyalty/internal/services/auth_service"
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

// @Summary Создание нового выражения
// @Description Создает новое выражение на сервере
// @Tags Calculations
// @Accept json
// @Produce json
// @Param body body Request true "Ready"
// @Success 200 {object} Response
// @Router /auth/ready [get]
// @Security BearerAuth
func (m *HealthHandlers) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responseError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx, cancel := context.WithTimeoutCause(r.Context(), 300*time.Millisecond, errors.New("readinessProbe timeout"))
	defer cancel()

	err := m.authservice.HealthCheck(ctx)

	if err != nil {
		responseError(w, r, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusOK)
	responseHealth(w, r, http.StatusOK, "ready")
}

func (m *HealthHandlers) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responseError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	responseHealth(w, r, http.StatusOK, "alive")
}

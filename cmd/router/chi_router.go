package router

import (
	"compress/gzip"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	handlersV1 "github.com/AlexBlackNn/authloyalty/internal/handlers/v1"
	customMiddleware "github.com/AlexBlackNn/authloyalty/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"log/slog"
	"time"
)

func NewChiRouter(
	cfg *config.Config,
	log *slog.Logger,
	authHandlerV1 handlersV1.AuthHandlers,
	healthHandlerV1 handlersV1.HealthHandlers,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	//	Rate limit by IP and URL path (aka endpoint)
	router.Use(httprate.Limit(
		cfg.RateLimit, // requests
		time.Second,   // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))
	router.Use(customMiddleware.Logger(log))
	router.Use(customMiddleware.GzipDecompressor(log))
	router.Use(customMiddleware.GzipCompressor(log, gzip.BestCompression))

	router.Use(middleware.Recoverer)

	router.Route("/auth", func(r chi.Router) {
		r.Get("/ready", healthHandlerV1.ReadinessProbe)
		r.Get("/healthz", healthHandlerV1.LivenessProbe)
		r.Post("/login", authHandlerV1.Login)
		r.Post("/logout", authHandlerV1.Logout)
		r.Post("/registration", authHandlerV1.Register)
	})
	return router
}

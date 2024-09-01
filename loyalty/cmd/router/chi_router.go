package router

import (
	"compress/gzip"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	v1 "github.com/AlexBlackNn/authloyalty/loyalty/internal/handlershttp/http/v1"

	_ "github.com/AlexBlackNn/authloyalty/loyalty/cmd/docs"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	customMiddleware "github.com/AlexBlackNn/authloyalty/loyalty/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewChiRouter(
	cfg *config.Config,
	log *slog.Logger,
	loyaltyhHandlerV1 v1.LoyaltyHandlers,
	healthHandlerV1 v1.HealthHandlers,
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
	router.Use(middleware.Recoverer)

	router.Route("/loyalty", func(r chi.Router) {
		r.Use(customMiddleware.GzipDecompressor(log))
		r.Use(customMiddleware.GzipCompressor(log, gzip.BestCompression))
		r.Get("/{uuid}", loyaltyhHandlerV1.GetLoyalty)
		r.Post("/", loyaltyhHandlerV1.AddLoyalty)
		r.Get("/ready", healthHandlerV1.ReadinessProbe)
		r.Get("/healthz", healthHandlerV1.LivenessProbe)

	})
	router.Route("/", func(r chi.Router) {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8001/swagger/doc.json"),
		))
		r.Get("/metrics", promhttp.Handler().ServeHTTP)
	})

	return router
}

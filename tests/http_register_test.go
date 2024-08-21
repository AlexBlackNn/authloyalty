package tests

import (
	"context"
	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/cmd/sso/router"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	"github.com/AlexBlackNn/authloyalty/pkg/storage/patroni"
	"github.com/AlexBlackNn/authloyalty/pkg/storage/redissentinel"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type AuthSuite struct {
	suite.Suite
	application *serverhttp.App
	client      http.Client
	srv         *httptest.Server
}

func (ms *AuthSuite) SetupSuite() {
	var err error

	cfg := config.MustLoadByPath("../config/local.yaml")
	log := logger.New(cfg.Env)

	userStorage, err := patroni.New(cfg)
	ms.Suite.NoError(err)

	tokenStorage, err := redissentinel.New(cfg)
	ms.Suite.NoError(err)

	producer, err := broker.New(cfg)
	ms.Suite.NoError(err)

	authService := authservice.New(
		cfg,
		log,
		userStorage,
		tokenStorage,
		producer,
	)

	// http server
	ms.application, err = serverhttp.New(cfg, log, authService)
	ms.Suite.NoError(err)
	ms.client = http.Client{Timeout: 3 * time.Second}
}

func (ms *AuthSuite) BeforeTest(suiteName, testName string) {
	// Starts server with first random port.
	ms.srv = httptest.NewServer(router.NewChiRouter(
		ms.application.Cfg,
		ms.application.Log,
		ms.application.HandlersV1,
		ms.application.HealthChecker,
	))
}

func (ms *AuthSuite) AfterTest(suiteName, testName string) {
	ms.srv = nil
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}

func (ms *AuthSuite) TestServerHappyPathV1() {

	type Want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name string
		url  string
		want Want
	}{
		{
			name: "gauge with value 10.3",
			url:  "/update/gauge/Lookups/10.3",
			want: Want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "counter with value 10",
			url:  "/update/counter/PoolCount/10",
			want: Want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	// stop server when tests finished
	defer ms.srv.Close()

	for _, tt := range tests {
		ms.Run(tt.name, func() {
			url := ms.srv.URL + tt.url
			request, err := http.NewRequest(http.MethodPost, url, nil)
			ms.NoError(err)
			res, err := ms.client.Do(request)
			ms.NoError(err)
			ms.Equal(tt.want.code, res.StatusCode)
			defer res.Body.Close()
			ms.Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func (ms *MetricsSuite) TestServerGetMetricHappyPathGauge() {
	type Want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name        string
		url         string
		metricType  string
		metricName  string
		metricValue float64
		want        Want
	}{
		{
			name:        "gauge with value 10.3",
			url:         "/value/gauge/test_gauge",
			metricType:  "gauge",
			metricName:  "test_gauge",
			metricValue: 10.3,
			want: Want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				response:    "10.3",
			},
		},
		{
			name:        "gauge with value 20.3",
			url:         "/value/gauge/test_gauge",
			metricType:  "gauge",
			metricName:  "test_gauge",
			metricValue: -20.3,
			want: Want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				response:    "-20.3",
			},
		},
	}
	// Stop server when tests finished.
	defer ms.srv.Close()

	for _, tt := range tests {
		ms.Run(tt.name, func() {
			metric := &models.Metric[float64]{Type: tt.metricType, Name: tt.metricName, Value: tt.metricValue}
			err := ms.application.MetricsService.UpdateMetricValue(context.Background(), metric)
			ms.NoError(err)
			url := ms.srv.URL + tt.url
			request, err := http.NewRequest(http.MethodGet, url, nil)
			ms.NoError(err)
			res, err := ms.client.Do(request)
			ms.NoError(err)
			ms.Equal(tt.want.code, res.StatusCode)
			bodyBytes, err := io.ReadAll(res.Body)
			ms.NoError(err)
			ms.Equal(tt.want.response, string(bodyBytes))
			defer res.Body.Close()
			ms.Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func (ms *MetricsSuite) TestServerGetMetricHappyPathCounter() {
	type Want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name        string
		url         string
		metricType  string
		metricName  string
		metricValue uint64
		want        Want
	}{
		{
			name:        "counter with value 10",
			url:         "/value/counter/test_counter",
			metricType:  "counter",
			metricName:  "test_counter",
			metricValue: 10,
			want: Want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				response:    "10",
			},
		},
	}
	// Stop server when tests finished.
	defer ms.srv.Close()

	for _, tt := range tests {
		ms.Run(tt.name, func() {
			metric := &models.Metric[uint64]{Type: tt.metricType, Name: tt.metricName, Value: tt.metricValue}
			err := ms.application.MetricsService.UpdateMetricValue(context.Background(), metric)
			ms.NoError(err)
			url := ms.srv.URL + tt.url
			request, err := http.NewRequest(http.MethodGet, url, nil)
			ms.NoError(err)
			res, err := ms.client.Do(request)
			ms.NoError(err)
			ms.Equal(tt.want.code, res.StatusCode)
			bodyBytes, err := io.ReadAll(res.Body)
			ms.NoError(err)
			ms.Equal(tt.want.response, string(bodyBytes))
			defer res.Body.Close()
			ms.Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func (ms *MetricsSuite) TestNegativeCasesMetrics() {

	type Want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name string
		url  string
		want Want
	}{
		{
			name: "metric name absent",
			url:  "/update/counter/10",
			want: Want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "metric wrong type value",
			url:  "/update/counter/PoolCount/test",
			want: Want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "metric wrong metric type",
			url:  "/update/histogram/PoolCount/10",
			want: Want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	defer ms.srv.Close()
	for _, tt := range tests {
		ms.Run(tt.name, func() {
			request, err := http.NewRequest(http.MethodPost, ms.srv.URL+tt.url, nil)
			ms.NoError(err)
			res, err := ms.client.Do(request)
			ms.NoError(err)
			ms.Equal(tt.want.code, res.StatusCode)
			defer res.Body.Close()
			ms.Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func (ms *MetricsSuite) TestNegativeCasesRequestMethods() {

	type Want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name   string
		url    string
		method string
		want   Want
	}{
		{
			name:   "get request",
			url:    "/update/counter/10",
			method: http.MethodGet,
			want: Want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "delete request",
			url:    "/update/counter/10",
			method: http.MethodDelete,
			want: Want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "put request",
			url:    "/update/counter/10",
			method: http.MethodPut,
			want: Want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "patch request",
			url:    "/update/counter/10",
			method: http.MethodPatch,
			want: Want{
				code: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		ms.Run(tt.name, func() {
			request, err := http.NewRequest(tt.method, ms.srv.URL+tt.url, nil)
			ms.NoError(err)
			res, err := ms.client.Do(request)
			ms.NoError(err)
			defer res.Body.Close()
			ms.Equal(tt.want.code, res.StatusCode)
		})
	}
}

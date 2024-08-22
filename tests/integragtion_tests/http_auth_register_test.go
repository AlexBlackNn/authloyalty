package integragtion_tests

import (
	"bytes"
	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/cmd/sso/router"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	"github.com/AlexBlackNn/authloyalty/pkg/storage/patroni"
	"github.com/AlexBlackNn/authloyalty/pkg/storage/redissentinel"
	"github.com/AlexBlackNn/authloyalty/tests/integragtion_tests/common"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
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

	cfg := config.MustLoadByPath("../../config/local.yaml")
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

func (ms *AuthSuite) TestHttpServerRegisterHappyPath() {
	type Want struct {
		code        int
		response    domain.Response
		contentType string
	}

	regBody := domain.Register{
		Email:    gofakeit.Email(),
		Password: common.RandomFakePassword(),
		Name:     gofakeit.Name(),
		Birthday: gofakeit.Date().Format("2006-01-02"),
	}
	reqJSON, err := regBody.MarshalJSON()
	ms.NoError(err)

	test := struct {
		name string
		url  string
		body []byte
		want Want
	}{
		name: "user registration",
		url:  "/auth/registration",
		body: reqJSON,
		want: Want{
			code:        http.StatusCreated,
			contentType: "application/json",
			response:    domain.Response{Status: "Success"},
		},
	}
	// stop server when tests finished
	defer ms.srv.Close()

	ms.Run(test.name, func() {
		url := ms.srv.URL + test.url
		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(test.body))
		ms.NoError(err)
		registerTime := time.Now() // to check token expiration time
		res, err := ms.client.Do(request)
		ms.NoError(err)
		ms.Equal(test.want.code, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		ms.NoError(err)

		var response domain.Response
		err = response.UnmarshalJSON(body)
		ms.NoError(err)
		ms.Equal(test.want.response.Status, response.Status)

		tokenParsed, err := jwt.Parse(response.AccessToken, func(token *jwt.Token) (any, error) {
			return []byte(ms.application.Cfg.ServiceSecret), nil
		})
		ms.NoError(err)

		// check validation
		claims, ok := tokenParsed.Claims.(jwt.MapClaims)
		ms.Suite.True(ok)
		// checking token expiration time might be only approximate
		const deltaSeconds = 1
		ms.Suite.InDelta(registerTime.Add(ms.application.Cfg.AccessTokenTtl).Unix(), claims["exp"].(float64), deltaSeconds)
		defer res.Body.Close()
		ms.Equal(test.want.contentType, res.Header.Get("Content-Type"))
	})
}

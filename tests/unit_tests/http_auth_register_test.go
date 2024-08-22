package unit_tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/cmd/sso/router"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/internal/dto"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	"github.com/AlexBlackNn/authloyalty/tests/unit_tests/mocks"
	gofakeit "github.com/brianvoe/gofakeit/v6"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
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

	ctrl := gomock.NewController(ms.T())
	defer ctrl.Finish()

	userStorageMock := mocks.NewMockUserStorage(ctrl)
	userStorageMock.EXPECT().
		SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), "79d3ac44-5857-4185-ba92-1a224fbacb51", nil).
		AnyTimes()

	passHash, err := bcrypt.GenerateFromPassword(
		[]byte("test"), bcrypt.DefaultCost,
	)

	user := domain.User{
		ID:       "79d3ac44-5857-4185-ba92-1a224fbacb51",
		Email:    "test@test.com",
		PassHash: passHash,
		IsAdmin:  false,
	}
	userStorageMock.EXPECT().
		GetUserByEmail(gomock.Any(), gomock.Any()).
		Return(context.Background(), user, nil).
		AnyTimes()

	userStorageMock.EXPECT().
		UpdateSendStatus(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), nil).
		AnyTimes()

	brokerMock := mocks.NewMockGetResponseChanSender(ctrl)

	brokerMock.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), nil).
		AnyTimes()

	brokerMock.EXPECT().
		GetResponseChan().
		Return(make(chan *broker.Response)).
		AnyTimes()

	tokenStorageMock := mocks.NewMockTokenStorage(ctrl)
	authService := authservice.New(
		cfg,
		log,
		userStorageMock,
		tokenStorageMock,
		brokerMock,
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
		response    dto.Response
		contentType string
	}

	regBody := dto.Register{
		Email:    "test@test.com",
		Password: "test",
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
			response:    dto.Response{Status: "Success"},
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

		var response dto.Response
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

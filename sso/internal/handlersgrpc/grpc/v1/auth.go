package v1

import (
	"context"
	"errors"
	log "log/slog"

	"github.com/AlexBlackNn/authloyalty/sso/internal/domain"
	"github.com/AlexBlackNn/authloyalty/sso/internal/dto"
	"github.com/AlexBlackNn/authloyalty/sso/internal/storage"

	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/sso/internal/services/authservice"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authService interface {
	Login(
		ctx context.Context,
		reqData *dto.Login,
	) (userWithTokens *domain.UserWithTokens, err error)
	Register(
		ctx context.Context,
		reqData *dto.Register,
	) (ctxOut context.Context, userWithTokens *domain.UserWithTokens, err error)
	Logout(
		ctx context.Context,
		reqData *dto.Logout,
	) (success bool, err error)
	IsAdmin(
		ctx context.Context,
		userID string,
	) (success bool, err error)
	Validate(
		ctx context.Context,
		token string,
	) (success bool, err error)
	Refresh(
		ctx context.Context,
		reqData *dto.Refresh,
	) (userWithTokens *domain.UserWithTokens, err error)
}

// serverAPI TRANSPORT layer
type serverAPI struct {
	// provides ability to work even without service interface realisation
	ssov1.UnimplementedAuthServer
	// service layer
	auth   authService
	tracer trace.Tracer
}

func Register(gRPC *grpc.Server, auth authService) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth, tracer: otel.Tracer("sso service")})
}

const (
	emptyId = ""
)

//realisation of transport layer interface
// see sso_grpc.pb.go ssov1.UnimplementedAuthServer

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	ctx, err := getContextWithTraceId(ctx)
	if err != nil {
		log.Warn(err.Error())
	}
	if err = validateLogin(req); err != nil {
		return nil, err
	}

	userWithTokens, err := s.auth.Login(
		ctx, &dto.Login{Email: req.GetEmail(), Password: req.GetPassword()},
	)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.LoginResponse{
		AccessToken:  userWithTokens.AccessToken,
		RefreshToken: userWithTokens.RefreshToken,
	}, nil
}

func (s *serverAPI) Refresh(
	ctx context.Context,
	req *ssov1.RefreshRequest,
) (*ssov1.RefreshResponse, error) {

	ctx, err := getContextWithTraceId(ctx)
	if err != nil {
		log.Warn(err.Error())
	}

	userWithTokens, err := s.auth.Refresh(
		ctx, &dto.Refresh{Token: req.GetRefreshToken()},
	)
	if err != nil {
		if errors.Is(err, authservice.ErrTokenWrongType) {
			return nil, status.Error(codes.InvalidArgument, "Provide valid refresh token")
		}
		if errors.Is(err, authservice.ErrTokenRevoked) {
			return nil, status.Error(codes.Unauthenticated, "Provide valid refresh token")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RefreshResponse{
		AccessToken:  userWithTokens.AccessToken,
		RefreshToken: userWithTokens.RefreshToken,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	ctx, err := getContextWithTraceId(ctx)
	if err != nil {
		log.Warn(err.Error())
	}

	if err = validateRegister(req); err != nil {
		return nil, err
	}
	ctx, userWithTokens, err := s.auth.Register(
		ctx, &dto.Register{Email: req.GetEmail(), Password: req.GetPassword()},
	)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(
				codes.AlreadyExists, "user already exists",
			)
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.RegisterResponse{
		UserId: userWithTokens.ID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	ctx, err := getContextWithTraceId(ctx)
	if err != nil {
		log.Warn(err.Error())
	}

	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}
	// call IsAdmin from service layer
	IsAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		// TODO: add error processing depends on the type of error
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: IsAdmin,
	}, nil
}

func (s *serverAPI) Logout(
	ctx context.Context,
	req *ssov1.LogoutRequest,
) (*ssov1.LogoutResponse, error) {
	ctx, err := getContextWithTraceId(ctx)
	if err != nil {
		log.Warn(err.Error())
	}
	success, err := s.auth.Logout(
		ctx, &dto.Logout{Token: req.GetToken()},
	)
	if err != nil {
		// TODO: add error processing depends on the type of error
		return nil, status.Error(codes.InvalidArgument, "bad token")
	}
	return &ssov1.LogoutResponse{Success: success}, nil
}

func (s *serverAPI) Validate(
	ctx context.Context,
	req *ssov1.ValidateRequest,
) (*ssov1.ValidateResponse, error) {
	ctx, err := getContextWithTraceId(ctx)
	if err != nil {
		log.Warn(err.Error())
	}
	success, err := s.auth.Validate(ctx, req.GetToken())
	if err != nil {
		// TODO: add error processing depends on the type of error
		return nil, err
	}
	return &ssov1.ValidateResponse{Success: success}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	//TODO: use special packet for data validation
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	//TODO: use special packet for data validation
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	//TODO: use special packet for data validation
	if req.GetUserId() == emptyId {
		return status.Error(codes.InvalidArgument, "userid is required")
	}
	return nil
}

func getContextWithTraceId(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, errors.New("metadata is absent in request")
	}
	traceIdString, ok := md["x-trace-id"]
	if !ok {
		return ctx, errors.New("x-trace-id is absent")
	}
	traceId, err := trace.TraceIDFromHex(traceIdString[0])
	if err != nil {
		return ctx, err
	}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceId,
	})
	return trace.ContextWithSpanContext(ctx, spanContext), nil
}

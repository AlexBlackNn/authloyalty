package ssoclient

import (
	"context"
	"net"

	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func grpcAddress() string {
	//TODO: sso migt be localhost in local run. Move to cfg
	return net.JoinHostPort("localhost", "44044")
}

type SSOClient struct {
	AuthClient ssov1.AuthClient
}

func New() (*SSOClient, error) {
	grpcClient, err := grpc.NewClient(
		grpcAddress(),
		//use insecure connection during test
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}
	authClient := ssov1.NewAuthClient(grpcClient)
	return &SSOClient{AuthClient: authClient}, nil
}

func (sc *SSOClient) IsJWTValid(ctx context.Context, tracer trace.Tracer, token string) bool {
	ctx, span := tracer.Start(ctx, "sso client: IsJWTValid",
		trace.WithAttributes(attribute.String("operation", "IsJWTValid")))
	defer span.End()

	respIsValid, err := sc.AuthClient.Validate(ctx, &ssov1.ValidateRequest{Token: token})
	if err != nil {
		return false
	}
	return respIsValid.GetSuccess()
}

func (sc *SSOClient) IsAdmin(ctx context.Context, tracer trace.Tracer, uuid string) bool {
	ctx, span := tracer.Start(ctx, "sso client: IsAdmin",
		trace.WithAttributes(attribute.String("operation", "IsAdmin")))
	defer span.End()

	respIsValid, err := sc.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{UserId: uuid})
	if err != nil {
		return false
	}
	return respIsValid.GetIsAdmin()
}

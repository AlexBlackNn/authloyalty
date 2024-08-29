package lib

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func grpcAddress() string {
	//TODO: sso migt be localhost in local run. Move to cfg
	return net.JoinHostPort("localhost", "44044")
}

func JWTCheck(ctx context.Context, tracer trace.Tracer, token string) bool {
	ctx, span := tracer.Start(ctx, "service layer: JWTCheck",
		trace.WithAttributes(attribute.String("handler", "JWTCheck")))
	defer span.End()

	traceId := fmt.Sprintf("%s", span.SpanContext().TraceID())
	ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", traceId)

	// create grpc_transport client
	cc, err := grpc.DialContext(
		context.Background(),
		grpcAddress(),
		//use insecure connection during test
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)

	if err != nil {
		return false
	}

	authClient := ssov1.NewAuthClient(cc)

	respIsValid, err := authClient.Validate(ctx, &ssov1.ValidateRequest{Token: token})
	return respIsValid.GetSuccess()
}

func JWTParse(tokenString string) (string, string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Fatal(err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Fatal("invalid claims format")
	}

	for key, value := range claims {
		fmt.Printf("%s = %v\n", key, value)
	}

	userName, ok := claims["email"].(string)
	if !ok {
		return "", "", errors.New("email not found")
	}
	userId, ok := claims["uid"].(string)
	if !ok {
		return "", "", errors.New("uid not found")
	}

	return userId, userName, nil
}

package lib

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Interceptor struct {
	tracer trace.Tracer
}

func (i *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, span := i.tracer.Start(ctx, "transport layer: "+info.FullMethod)
		defer span.End()
		resp, err := handler(ctx, req)
		return resp, err
	}
}

func NewInterceptor(tracer trace.Tracer) *Interceptor {
	return &Interceptor{tracer: tracer}
}

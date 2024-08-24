package interceptors

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Tracing struct {
	tracer trace.Tracer
}

func (i *Tracing) GetInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, span := i.tracer.Start(ctx, "transport layer: "+info.FullMethod)
		defer span.End()
		resp, err := handler(ctx, req)
		return resp, err
	}
}

func NewTracing(tracer trace.Tracer) *Tracing {
	return &Tracing{tracer: tracer}
}

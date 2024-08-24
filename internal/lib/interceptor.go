package lib

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Interceptor struct {
	tracer trace.Tracer
}

func (i *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Запись имени метода в качестве атрибута трассировки
		fmt.Println("111111111111111111111111111111111111111")
		ctx, span := i.tracer.Start(ctx, "transport layer1111: "+info.FullMethod)
		defer span.End()
		fmt.Println("222222222222222222222")
		// Вызов обработчика
		resp, err := handler(ctx, req)
		fmt.Println("333333333333333")

		return resp, err
	}
}

func NewInterceptor(tracer trace.Tracer) *Interceptor {
	return &Interceptor{tracer: tracer}
}

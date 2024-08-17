package main

import (
	"context"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/commands/proto/registration.v1/registration.v1"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/pkg/tracing"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	log "log/slog"
	"os"
	"time"
)

type OTelInterceptor struct {
	tracer     trace.Tracer
	fixedAttrs []attribute.KeyValue
}

// NewOTelInterceptor processes span for intercepted messages and add some
// headers with the span data.
func NewOTelInterceptor(groupID string) *OTelInterceptor {
	oi := OTelInterceptor{}
	oi.tracer = otel.Tracer("consumer")

	oi.fixedAttrs = []attribute.KeyValue{
		semconv.MessagingSystemKafka,
		semconv.MessagingKafkaConsumerGroup(groupID),
		semconv.NetworkTransportTCP,
	}
	return &oi
}

func (oi *OTelInterceptor) OnConsume(ctx context.Context, kafkaHeaders []kafka.Header) context.Context {
	headers := propagation.MapCarrier{}

	for _, recordHeader := range kafkaHeaders {
		headers[recordHeader.Key] = string(recordHeader.Value)
	}

	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, headers)

	ctx, span := oi.tracer.Start(
		ctx,
		"tracer consumer1",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(oi.fixedAttrs...),
		trace.WithAttributes(
			semconv.MessagingDestinationName("registration"),
		),
	)
	span.End()
	ctx, span = oi.tracer.Start(
		ctx,
		"tracer consumer2",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(oi.fixedAttrs...),
		trace.WithAttributes(
			semconv.MessagingDestinationName("registration"),
		),
	)
	span.End()
	time.Sleep(1 * time.Second)
	return ctx
}

func main() {

	cfg := config.New()
	_, err := tracing.Init("consumer", cfg)
	if err != nil {
		log.Error(err.Error())
	}
	var customTracer = otel.Tracer("consumer")
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9094",
		"group.id":           "1",
		"session.timeout.ms": 6000,
		"auto.offset.reset":  "earliest"})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	client, err := schemaregistry.NewClient(schemaregistry.NewConfig("http://localhost:8081"))

	if err != nil {
		fmt.Printf("Failed to create schema registry client: %s\n", err)
		os.Exit(1)
	}

	deser, err := protobuf.NewDeserializer(client, serde.ValueSerde, protobuf.NewDeserializerConfig())

	if err != nil {
		fmt.Printf("Failed to create deserializer: %s\n", err)
		os.Exit(1)
	}

	// Register the Protobuf type so that Deserialize can be called.
	// An alternative is to pass a pointer to an instance of the Protobuf type
	// to the DeserializeInto method.
	deser.ProtoRegistry.RegisterMessage((&registration_v1.RegistrationMessage{}).ProtoReflect().Type())
	//TODO: registration should be got from config
	err = c.Subscribe("registration", nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe to topics: %s\n", err)
		os.Exit(1)
	}
	for {
		ev := c.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			ctx := context.Background()
			ctx, span := customTracer.Start(ctx, "kafka_message_processing")
			defer span.End()

			value, err := deser.Deserialize(*e.TopicPartition.Topic, e.Value)
			if err != nil {
				fmt.Printf("Failed to deserialize payload: %s\n", err)
			} else {
				fmt.Printf("%% Message on %s:\n%+v\n", e.TopicPartition, value)
			}
			if e.Headers != nil {
				fmt.Printf("%% Headers: %v\n", e.Headers)

				headers := propagation.MapCarrier{}

				for _, recordHeader := range e.Headers {
					headers[recordHeader.Key] = string(recordHeader.Value)
				}

				propagator := otel.GetTextMapPropagator()
				ctx = propagator.Extract(ctx, headers)

				ctx, span = customTracer.Start(
					ctx,
					"tracer consumer1",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingDestinationName("registration"),
					),
				)
				span.End()
				ctx, span = customTracer.Start(
					ctx,
					"tracer consumer2",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingDestinationName("registration"),
					),
				)
				span.End()
				fmt.Println("1111111111111111111111", ctx)
				ctx, span = customTracer.Start(
					ctx,
					"tracer consumer3",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingDestinationName("registration"),
					),
				)
				span.End()
				foo(ctx)
				ctx, span = customTracer.Start(
					ctx,
					"tracer consumer4",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingDestinationName("registration"),
					),
				)
				span.End()

			}
		case kafka.Error:
			// Errors should generally be considered
			// informational, the client will try to
			// automatically recover.
			fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
		default:
			fmt.Printf("Ignored %v\n", e)
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}

func foo(ctx context.Context) context.Context {
	var customTracer = otel.Tracer("consumer")
	fmt.Printf("Hello\n")
	fmt.Println("1111111111111111111111", ctx)
	ctx, span := customTracer.Start(
		ctx,
		"foo",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			semconv.MessagingDestinationName("registration"),
		),
	)
	span.End()
	return ctx
}

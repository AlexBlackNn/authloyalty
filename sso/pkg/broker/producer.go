package broker

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/sso/internal/config"
	"github.com/AlexBlackNn/authloyalty/sso/pkg/tracing/otelconfluent"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/protobuf"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type Broker struct {
	producer     *otelconfluent.Producer
	serializer   serde.Serializer
	ResponseChan chan *Response
}

type Response struct {
	UserUUID string
	Err      error
}

var FlushBrokerTimeMs = 100
var KafkaError = errors.New("kafka broker failed")
var tracer = otel.Tracer(
	"sso service",
	trace.WithInstrumentationVersion(contrib.SemVersion()),
)

// New returns kafka producer with schema registry
func New(cfg *config.Config) (*Broker, error) {
	confluentProducer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Kafka.KafkaURL})
	if err != nil {
		return nil, err
	}
	p := otelconfluent.NewProducerWithTracing(
		confluentProducer,
		tracer,
	)
	c, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.Kafka.SchemaRegistryURL))
	if err != nil {
		return nil, err
	}
	s, err := protobuf.NewSerializer(c, serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, err
	}

	kafkaResponseChan := make(chan *Response)

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch e := e.(type) {
			// https://github.com/confluentinc/confluent-kafka-go/blob/master/examples/producer_example/producer_example.go
			case *kafka.Message:
				// The message delivery report, indicating success or
				// permanent failure after retries have been exhausted.
				// Application level retries won't help since the client
				// is already configured to do that.
				if e.TopicPartition.Error != nil {
					kafkaResponseChan <- &Response{
						UserUUID: string(e.Key),
						Err:      e.TopicPartition.Error,
					}
					continue
				}
				kafkaResponseChan <- &Response{
					UserUUID: string(e.Key),
					Err:      nil,
				}
			case kafka.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				//
				// These errors should generally be considered informational
				// as the underlying client will automatically try to
				// recover from any errors encountered, the application
				// does not need to take action on them.
				kafkaResponseChan <- &Response{
					UserUUID: "",
					Err:      fmt.Errorf("kafka general error %w - %v", KafkaError, e),
				}
			}
		}
	}()

	return &Broker{
			producer:     p,
			serializer:   s,
			ResponseChan: kafkaResponseChan,
		},
		nil
}

// Close closes serialization agent and kafka producer
func (b *Broker) Close() {
	b.serializer.Close()
	//https://docs.confluent.io/platform/current/clients/confluent-kafka-go/index.html#hdr-Producer
	//* When done producing messages it's necessary  to make sure all messages are
	//indeed delivered to the broker (or failed),
	//because this is an asynchronous client so some messages may be
	//lingering in internal channels or transmission queues.
	//Call the convenience function `.Flush()` will block code until all
	//message deliveries are done or the provided timeout elapses.
	b.producer.Flush(FlushBrokerTimeMs)
	b.producer.Close()
}

// GetResponseChan returns channel to get messages send status
func (b *Broker) GetResponseChan() chan *Response {
	return b.ResponseChan
}

// Send sends serialized message to kafka using schema registry
func (b *Broker) Send(ctx context.Context, msg proto.Message, topic string, key string) (context.Context, error) {
	ctx, span := tracer.Start(
		ctx, "transfer layer Kafka: Serialize message",
		trace.WithAttributes(attribute.String("transfer transfer", "Send")),
	)
	payload, err := b.serializer.Serialize(topic, msg)
	if err != nil {
		return ctx, err
	}
	span.End()
	ctx, span = tracer.Start(
		ctx, "transfer layer Kafka: Send message",
		trace.WithAttributes(attribute.String("transfer transfer", "Send")),
	)
	defer span.End()
	headers := []kafka.Header{{Key: "request-Id", Value: []byte("header values are binary")}}

	// add span to headers to send via kafka
	headers, span = createProducerSpan(ctx, headers)
	defer span.End()

	if ctx, err = b.producer.Produce(ctx, &kafka.Message{
		Key:            []byte(key),
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          payload,
		Headers:        headers,
	}, nil); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func createProducerSpan(ctx context.Context, headers []kafka.Header) ([]kafka.Header, trace.Span) {
	ctx, span := tracer.Start(
		ctx,
		"transfer layer Kafka: to target services",
		trace.WithAttributes(
			semconv.PeerService("kafka"),
			semconv.NetworkTransportTCP,
			semconv.MessagingSystemKafka,
			semconv.MessagingDestinationName("registration"),
		),
	)

	carrier := propagation.MapCarrier{}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, carrier)

	for key, value := range carrier {
		headers = append(headers, kafka.Header{Key: key, Value: []byte(value)})
	}

	return headers, span
}

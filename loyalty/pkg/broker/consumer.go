package broker

import (
	"context"
	"errors"
	"fmt"
	"os"

	registrationv1 "github.com/AlexBlackNn/authloyalty/commands/proto/registration.v1/registration.v1"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/protobuf"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type MessageReceived struct {
	Message *kafka.Message
	Err     error
}

type Broker struct {
	consumer     *kafka.Consumer
	deserializer serde.Deserializer
	MessageChan  chan *MessageReceived
}

var FlushBrokerTimeMs = 100
var KafkaError = errors.New("kafka broker failed")
var tracer = otel.Tracer(
	"loyalty service",
	trace.WithInstrumentationVersion(contrib.SemVersion()),
)

// New returns kafka consumer with schema registry
func New(cfg *config.Config) (*Broker, error) {
	confluentConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.Kafka.KafkaURL,
		"group.id":           "1",
		"session.timeout.ms": 6000,
		"auto.offset.reset":  "earliest"})
	if err != nil {
		return nil, err
	}

	//p := otelconfluent.NewConsumerWithTracing(
	//	confluentConsumer,
	//	tracer,
	//)

	MessageChan := make(chan *MessageReceived)

	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.Kafka.SchemaRegistryURL))
	if err != nil {
		return nil, err
	}

	deser, err := protobuf.NewDeserializer(client, serde.ValueSerde, protobuf.NewDeserializerConfig())
	if err != nil {
		return nil, err
	}

	deser.ProtoRegistry.RegisterMessage((&registrationv1.RegistrationMessage{}).ProtoReflect().Type())
	//TODO: registration should be got from config
	err = confluentConsumer.Subscribe("registration", nil)

	broker := &Broker{
		consumer:     confluentConsumer,
		deserializer: deser,
		MessageChan:  MessageChan,
	}
	broker.Consume()
	return broker, nil

}

// GetResponseChan returns channel to get messages send status
func (b *Broker) GetMessageChan() chan *MessageReceived {
	return b.MessageChan
}

// Close closes deserialization agent and kafka consumer
func (b *Broker) Close() error {
	b.deserializer.Close()
	//https://docs.confluent.io/platform/current/clients/confluent-kafka-go/index.html#hdr-High_level_Consumer
	err := b.consumer.Close()
	if err != nil {
		return err
	}
	return nil
}

func (b *Broker) Consume() {
	for {
		//TODO: get from config
		ev := b.consumer.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			ctx := context.Background()
			ctx, span := tracer.Start(ctx, "kafka_message_processing")
			defer span.End()

			value, err := b.deserializer.Deserialize(*e.TopicPartition.Topic, e.Value)
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

				ctx, span = tracer.Start(
					ctx,
					"tracer consumer1",
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
}

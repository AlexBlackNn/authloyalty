package broker

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/protobuf"
	"google.golang.org/protobuf/proto"
	"time"
)

type Broker struct {
	producer     *kafka.Producer
	serializer   serde.Serializer
	ResponseChan chan *Response
}

type Response struct {
	userId string
	err    error
}

// NewProducer returns kafka producer with schema registry
func NewProducer(kafkaURL, srURL string) (*Broker, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaURL})
	if err != nil {
		return nil, err
	}
	c, err := schemaregistry.NewClient(schemaregistry.NewConfig(srURL))
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

			case *kafka.Message:
				// The message delivery report, indicating success or
				// permanent failure after retries have been exhausted.
				// Application level retries won't help since the client
				// is already configured to do that.
				fmt.Println("1111111111111111111111111111", string(e.Key))
				kafkaResponseChan <- &Response{userId: string(e.Key), err: nil}
			case kafka.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				//
				// These errors should generally be considered informational
				// as the underlying client will automatically try to
				// recover from any errors encountered, the application
				// does not need to take action on them.

				kafkaResponseChan <- &Response{userId: "", err: err}
			default:
				fmt.Printf("Ignored event: %s\n", e)
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
	b.producer.Close()
}

// GetResponseChan returns channel to get messages send status
func (b *Broker) GetResponseChan() chan *Response {
	return b.ResponseChan
}

// Send sends serialized message to kafka using schema registry
func (b *Broker) Send(msg proto.Message, topic string, key string) error {
	payload, err := b.serializer.Serialize(topic, msg)
	if err != nil {
		return err
	}
	if err = b.producer.Produce(&kafka.Message{
		Key:            []byte(key),
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          payload,
		Headers:        []kafka.Header{{Key: "request-Id", Value: []byte("header values are binary")}},
	}, nil); err != nil {
		if err.(kafka.Error).Code() == kafka.ErrQueueFull {
			// Broker queue is full, wait 1s for messages
			// to be delivered then try again.
			time.Sleep(time.Second)
			return err
		}
		return err
	}
	return nil
}

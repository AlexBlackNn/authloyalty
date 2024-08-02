package broker

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/protobuf"
	"google.golang.org/protobuf/proto"
	"strconv"
	"time"
)

type Producer struct {
	producer   *kafka.Producer
	serializer serde.Serializer
}

// NewProducer returns kafka producer with schema registry
func NewProducer(kafkaURL, srURL string) (*Producer, error) {
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

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {

			case *kafka.Message:
				// The message delivery report, indicating success or
				// permanent failure after retries have been exhausted.
				// Application level retries won't help since the client
				// is already configured to do that.
				m := ev
				user_id, err := strconv.Atoi(string(m.Key))
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("1111111111111 user_id", user_id)

				if m.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
				} else {
					fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
			case kafka.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				//
				// These errors should generally be considered informational
				// as the underlying client will automatically try to
				// recover from any errors encountered, the application
				// does not need to take action on them.
				fmt.Printf("Error: %v\n", ev)
			default:
				fmt.Printf("Ignored event: %s\n", ev)
			}
		}
	}()

	return &Producer{
		producer:   p,
		serializer: s,
	}, nil
}

// Stop stops serialization agent and kafka producer
func (s *Producer) Stop() {
	s.serializer.Close()
	s.producer.Close()
}

// Send sends serialized message to kafka using schema registry
func (p *Producer) Send(msg proto.Message, topic string, key string) error {
	payload, err := p.serializer.Serialize(topic, msg)
	if err != nil {
		return err
	}
	if err = p.producer.Produce(&kafka.Message{
		Key:            []byte(key),
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          payload,
		Headers:        []kafka.Header{{Key: "request-Id", Value: []byte("header values are binary")}},
	}, nil); err != nil {
		if err.(kafka.Error).Code() == kafka.ErrQueueFull {
			// Producer queue is full, wait 1s for messages
			// to be delivered then try again.
			time.Sleep(time.Second)
			return err
		}
		fmt.Printf("Failed to produce message: %v\n", err)
		return err
	}
	return nil
}

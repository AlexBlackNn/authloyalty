package broker

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	nullOffset = -1
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
func (p *Producer) Send(msg proto.Message, topic string) error {
	kafkaChan := make(chan kafka.Event)
	defer close(kafkaChan)
	payload, err := p.serializer.Serialize(topic, msg)
	if err != nil {
		return err
	}
	if err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          payload,
	}, kafkaChan); err != nil {
		return err
	}
	e := <-kafkaChan
	switch e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return err
	}
	return nil
}

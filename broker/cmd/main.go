package main

import (
	"fmt"
	"github.com/AlexBlackNn/authloyalty/broker"
	test "github.com/AlexBlackNn/authloyalty/broker/test.v1"
	"log"
)

const (
	topic = "topic.v1"
)

func main() {
	kafkaURL := "localhost:9094"
	schemaRegistryURL := "http://localhost:8081"
	producer, err := broker.NewProducer(kafkaURL, schemaRegistryURL)
	defer producer.Close()

	if err != nil {
		log.Fatal(err)
	}
	testMSG := test.TestMessage{Value: 42}
	offset, err := producer.ProduceMessage(&testMSG, topic)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
}

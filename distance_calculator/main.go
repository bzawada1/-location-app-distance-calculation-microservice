package main

import (
	"log"
)

const topic = "obudata"

func main() {
	svc := NewCalculatorService()
	kafkaConsumer, err := NewKafkaConsumer(topic, svc)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
}

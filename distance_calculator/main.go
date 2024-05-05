package main

import (
	"log"
)

const topic = "obudata"

func main() {
	svc := NewCalculatorService()
	logSvc := NewLogMiddleware(svc)
	kafkaConsumer, err := NewKafkaConsumer(topic, logSvc)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
}

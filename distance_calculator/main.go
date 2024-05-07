package main

import (
	"github.com/bzawada1/location-app-obu-service/aggregator/client"
	"log"
)

const (
	topic              = "obudata"
	aggregatorEndpoint = "http://127.0.0.1:3000/aggregate"
)

func main() {
	svc := NewCalculatorService()
	logSvc := NewLogMiddleware(svc)
	client := client.NewHTTPClient(aggregatorEndpoint)
	kafkaConsumer, err := NewKafkaConsumer(topic, logSvc, client)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bzawada1/location-app-obu-service/types"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/websocket"
)

var kafkaTopic = "obudata"

func main() {
	// Delivery report handler for produced messages

	receiver, err := NewDataReceiver()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ws", receiver.handleWs)
	http.ListenAndServe(":30000", nil)
}

type DataReceiver struct {
	msg  chan types.OBUData
	conn *websocket.Conn
	prod *kafka.Producer
}

func NewDataReceiver() (*DataReceiver, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		return nil, err
	}
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()
	return &DataReceiver{
		prod: p,
	}, nil
}

func (dataReceiver *DataReceiver) produceData(data types.OBUData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = dataReceiver.prod.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kafkaTopic,
			Partition: kafka.PartitionAny,
		},
		Value: b,
	}, nil)
	return err
}

func (dataReceiver *DataReceiver) handleWs(w http.ResponseWriter, r *http.Request) {
	u := websocket.Upgrader{
		ReadBufferSize:  1028,
		WriteBufferSize: 1028,
	}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	dataReceiver.conn = conn
	go dataReceiver.wsReceiveLoop()
}

func (dataReceiver *DataReceiver) wsReceiveLoop() {
	fmt.Println("New OBU client connected")
	for {
		data := types.OBUData{}
		if err := dataReceiver.conn.ReadJSON(&data); err != nil {
			log.Printf("read error: ", err)
		}
		fmt.Println("received OBU data: ", data.Long)

		if err := dataReceiver.produceData(data); err != nil {
			log.Fatal(err)
		}
	}
}

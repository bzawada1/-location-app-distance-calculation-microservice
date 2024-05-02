package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bzawada1/location-app-obu-service/types"
	"github.com/gorilla/websocket"
)

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
	prod DataProducer
}

func NewDataReceiver() (*DataReceiver, error) {

	var p DataProducer
	p, err := NewKafkaProducer("obudata")
	if err != nil {
		return nil, err
	}

	p = NewLogMiddleware(p)
	return &DataReceiver{
		msg:  make(chan types.OBUData, 128),
		prod: p,
	}, nil
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

		if err := dataReceiver.prod.ProduceData(data); err != nil {
			log.Fatal(err)
		}
	}
}

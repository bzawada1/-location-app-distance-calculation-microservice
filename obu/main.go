package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/bzawada1/location-app-obu-service/types"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const wsEndpoint = "ws://127.0.0.1:30000/ws"

var sendInterval = 1 * time.Second

func sendOBUData(connection *websocket.Conn, data types.OBUData) error {
	return connection.WriteJSON(data)
}

func genLatLong() (float64, float64) {
	return genCord(), genCord()
}
func genCord() float64 {
	n := float64(rand.Intn(100) + 1)
	f := rand.Float64()
	return n + f
}

func main() {
	obuIDS := generateOBUIDS(20)
	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}
	for {
		for i := 0; i < len(obuIDS); i++ {
			lat, long := genLatLong()
			data := types.OBUData{
				OBUID: obuIDS[i],
				Lat:   lat,
				Long:  long,
			}
			logrus.WithFields(logrus.Fields{
				"obuID": data.OBUID,
			}).Info("OBU: generating data and sending it to receiver")
			if err := sendOBUData(conn, data); err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(sendInterval)
	}
}

func generateOBUIDS(n int) []int {
	ids := make([]int, n)
	for i := 0; i < n; i++ {
		ids[i] = rand.Intn(99999999)
	}
	return ids
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

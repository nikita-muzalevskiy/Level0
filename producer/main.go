package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	clusterID := "test-cluster"
	clientID := "producer"
	channel := "wb-channel"

	stanConn, err := stan.Connect(clusterID, clientID)
	if err != nil {
		log.Fatal(err)
	}
	defer stanConn.Close()

	_, err = stanConn.Subscribe(channel, func(msg *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(msg.Data))
	})
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open("jsons\\model.json")
	byteValue, _ := io.ReadAll(file)
	stanConn.Publish(channel, byteValue)

	// Ждем, чтобы обработчик успел получить и распечатать сообщение
	time.Sleep(1 * time.Second)
}

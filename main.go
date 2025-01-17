package main

import (
	"context"
	"fmt"
	EventClient "go-lang-eventbridge/pkg/client"
	EventServer "go-lang-eventbridge/pkg/server"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const SERVER_PORT = 8080

func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventServer, err := EventServer.NewEventServer(SERVER_PORT)
	if err != nil {
		panic(err)
	}

	go eventServer.Run(ctx)
	time.Sleep(100 * time.Millisecond)

	var (
		success    int32
		failed     int32
		inProgress int32
	)

	// Smaller number to make it clearer
	numConnections := 20

	for i := 0; i < numConnections; i++ {
		go func(id int) {

			atomic.AddInt32(&inProgress, 1)
			currentInProgress := atomic.LoadInt32(&inProgress)

			connStart := time.Now()
			log.Printf("Starting connection %d (In progress: %d)", id, currentInProgress)

			client := EventClient.NewEventClient(SERVER_PORT)
			err := client.SendMessage("topic la la ha", fmt.Sprintf("Message pkg %d\n", id))

			latency := time.Since(connStart)
			atomic.AddInt32(&inProgress, -1)

			if err != nil {
				atomic.AddInt32(&failed, 1)
				log.Printf("Connection %d failed after %v: %v", id, latency, err)
			} else {
				atomic.AddInt32(&success, 1)
				log.Printf("Connection %d completed in %v", id, latency)
			}
		}(i)

		// Optional: Add small delay between connection attempts
		time.Sleep(100 * time.Millisecond)
	}

	<-sigChan
	log.Println("Shutting down")
	cancel()
}

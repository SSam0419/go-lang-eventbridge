package main

import (
	"context"
	eventbus "go-lang-eventbridge/pkg/event_bus"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	eventbus.InitEventBus()

	eventbus.Instance.AddEventListener("event 1", "callback-1", func(ctx context.Context, data interface{}) error {
		time.Sleep(time.Second * 5)
		log.Println("Testing callback-1")
		log.Println("any data ? -> ", data)
		return nil
	})
	eventbus.Instance.AddEventListener("event 2", "callback-2", func(ctx context.Context, data interface{}) error {
		log.Println("Testing callback-2")
		return nil
	})

	for i := 0; i <= 2000; i++ {
		go eventbus.Instance.TriggerEvent(context.Background(), "event 1")
	}

	eventbus.Instance.TriggerEventWithPayload(context.Background(), "event 1", "just string")
	eventbus.Instance.TriggerEventWithPayload(context.Background(), "event 1", map[string]interface{}{
		"testing": "hello world",
	})

	<-sigChan
}

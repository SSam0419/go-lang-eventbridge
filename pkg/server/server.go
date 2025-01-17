package EventServer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	Event "go-lang-eventbridge/pkg/event"
)

type ServerListener struct {
	identifier string
}

func NewServerListener(identifier string) *ServerListener {
	return &ServerListener{
		identifier: identifier,
	}
}

type EventServer struct {
	connection net.Listener
	listeners  []*ServerListener
	lock       sync.RWMutex
	done       chan struct{}
}

func NewEventServer(port int) (*EventServer, error) {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &EventServer{
		connection: conn,
		listeners:  make([]*ServerListener, 0)}, nil
}

func (es *EventServer) Run(ctx context.Context) error {
	defer es.connection.Close()

	// Create a channel for accepting connections
	connChan := make(chan net.Conn, 10)

	// Connection acceptor goroutine
	go func() {
		for {
			conn, err := es.connection.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Printf("Error accepting connection: %v", err)
				}
				return
			}
			go func() {
				// simulate traffic
				log.Println("the wait before")
				time.Sleep(time.Second)
				log.Println("the wait after")

				connChan <- conn
			}()

		}
	}()

	go func() {
		for {
			log.Println("iterating ..")
			select {
			case <-es.done:
				return
			case <-ctx.Done():
				log.Println("")
				return
			case conn := <-connChan:
				go func(c net.Conn) {
					defer c.Close()
					message, err := Event.DecodeMessage(c)
					if err != nil {
						log.Printf("Read error: %v", err)
						return
					}
					log.Printf("[%s] : %s", message.Topic, message.Payload)
				}(conn)
			}

		}

	}()

	<-ctx.Done()

	return ctx.Err()

}

func (es *EventServer) AddListeners(identifier string) error {
	newListener := NewServerListener(identifier)
	es.lock.Lock()
	es.listeners = append(es.listeners, newListener)
	es.lock.Unlock()
	return nil
}

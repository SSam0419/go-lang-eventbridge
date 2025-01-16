package EventServer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

type Event struct{}

type ServerListener struct {
	identifier       string
	subscribedEvents []Event
}

func NewServerListener(identifier string) *ServerListener {
	return &ServerListener{
		identifier:       identifier,
		subscribedEvents: make([]Event, 0),
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

	<-ctx.Done()

	return ctx.Err()
	for {
		select {

		case conn, err := es.connection.Accept():
			if err != nil {
				log.Println("Error receiving event : ", err.Error())
			}
			w := bufio.NewReader(conn)
			message, err := w.ReadString('\n')
			if err != nil {
				log.Println("Error reading message : ", err.Error())
			}

			log.Println("Server received message : ", message)

		case <-ctx.Done():
			log.Println("")
			return

		}
	}
}

func (es *EventServer) AddListeners(identifier string) error {
	newListener := NewServerListener(identifier)
	es.lock.Lock()
	es.listeners = append(es.listeners, newListener)
	es.lock.Unlock()
	return nil
}

func (es *EventServer) SendMessage() error {
	return nil
}

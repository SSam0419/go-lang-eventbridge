package EventServer

import (
	"context"
	"errors"
	"fmt"
	protocol "go-lang-eventbridge/pkg/protocol"
	"log"
	"net"
	"sync"
)

type EventServer struct {
	connection net.Listener
	lock       sync.RWMutex
	done       chan struct{}
	//address of the clients listening on a topic
	topicAudience map[string][]string
	// maximum number of connections allowed at a time
	maxConnection int
}

func NewEventServer(port int) (*EventServer, error) {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}
	return &EventServer{
		connection:    conn,
		maxConnection: 10,
		topicAudience: make(map[string][]string),
	}, nil
}

func (es *EventServer) Run(ctx context.Context) error {
	defer es.connection.Close()

	// Create a channel for accepting connections
	connChan := make(chan net.Conn, es.maxConnection)

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
				connChan <- conn
			}()

		}
	}()

	go func() {
		for {
			select {
			case <-es.done:
				return
			case <-ctx.Done():
				log.Println("")
				return
			case conn := <-connChan:
				go func(c net.Conn) {
					defer c.Close()
					from := c.RemoteAddr()
					log.Println("client from  ", from, " sent an message")
					message, err := protocol.DecodeMessageFromConn(c)
					if err != nil {
						log.Printf("Read error: %v", err)
						return
					}

					log.Printf("Server received [%s] -> [%s] : %s", message.Command, message.Topic, message.Payload)

					if message.Command == protocol.ListenTopicCommand {
						go es.RegisterClient(message.ClientAddr, message.Topic)
					}

				}(conn)
			}

		}

	}()

	<-ctx.Done()

	return ctx.Err()

}

func (es *EventServer) RegisterClient(addr string, topic string) error {
	log.Printf("client %s is registering to listen on topic %s \n", addr, topic)
	es.lock.Lock()
	es.topicAudience[topic] = append(es.topicAudience[topic], addr)
	es.lock.Unlock()

	log.Println("current topicAudience map -> ", es.topicAudience)
	return nil
}

func (es *EventServer) Close() {
	es.done <- struct{}{}
}

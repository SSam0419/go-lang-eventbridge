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

type ClientRecord struct {
}

type EventServer struct {
	connection    net.Listener
	clientRecords []ClientRecord
	lock          sync.RWMutex
	done          chan struct{}
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
					// clientConn, err := net.Dial("tcp", from.String())
					// if err != nil {
					// 	log.Println("Error trying to send response back to client, ", err.Error())
					// } else {
					// 	Event.SendMessage("server response", "received msg from client", clientConn)
					// }
					message, err := protocol.DecodeMessageFromConn(c)
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

func (es *EventServer) RegisterClient() error {

	return nil
}

package EventClient

import (
	"context"
	"fmt"
	"go-lang-eventbridge/pkg/protocol"
	"log"
	"net"
)

type EventClient struct {
	serverConn     net.Conn
	clientListener net.Listener
}

func NewEventClient(port int) *EventClient {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("Error while sending message to Event Server at port ", port)
	}

	// empty port so that it will choose a port automatically
	listener, err := net.Listen("tcp", "")
	if err != nil {
		log.Println("Error while initializing a listener for client, ", err.Error())
	} else {
		log.Println("Client is listening at, ", listener.Addr().String())
		// then register the client to the server

	}

	return &EventClient{
		serverConn:     conn,
		clientListener: listener,
	}
}

func (ec *EventClient) Accept(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("context cancelled")
			return
		default:
			conn, err := ec.clientListener.Accept()
			if err != nil {
				log.Println("Error while accepting connection, ", err.Error())
				msg, err := protocol.DecodeMessageFromConn(conn)
				if err != nil {
					log.Println("Error while decoding message, ", err.Error())
				}
				log.Println("Client received -> ", msg)
			}
		}
	}
}

func (ec *EventClient) SendMessage(topic string, payload string) error {
	clientAddr := ec.clientListener.Addr().String()
	encoded, _ := protocol.EncodeMessage(topic, clientAddr, []byte(payload))
	_, err := ec.serverConn.Write(encoded)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (ec *EventClient) ListenTopic(topic string) error {
	clientAddr := ec.clientListener.Addr().String()
	encoded, _ := protocol.EncodeListenTopicRequest(topic, clientAddr)
	_, err := ec.serverConn.Write(encoded)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

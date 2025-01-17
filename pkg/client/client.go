package EventClient

import (
	"fmt"
	Event "go-lang-eventbridge/pkg/event"
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
	return &EventClient{
		serverConn: conn,
	}
}

func (ec *EventClient) SendMessage(topic string, payload string) error {
	encoded := Event.EncodePayload(topic, payload)
	_, err := ec.serverConn.Write(encoded)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

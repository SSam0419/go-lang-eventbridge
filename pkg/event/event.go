package Event

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Event struct {
	topic   string
	payload string
}

func NewEvent(topic string, payload string) *Event {
	return &Event{
		topic:   topic,
		payload: payload,
	}
}

// return one []byte following the pattern of
// [length of topic (2 bytes) ][topic][length of payload (4 bytes) ][payload]
func EncodePayload(topic string, payload string) []byte {
	// Prepare topic bytes and its length prefix (2 bytes)
	topicBytes := []byte(topic)
	topicLength := make([]byte, 2)
	binary.BigEndian.PutUint16(topicLength, uint16(len(topicBytes)))

	// Prepare payload bytes and its length prefix (4 bytes)
	payloadBytes := []byte(payload)
	payloadLength := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadLength, uint32(len(payloadBytes)))

	// Calculate total size and create result slice
	totalSize := 2 + len(topicBytes) + 4 + len(payloadBytes)
	result := make([]byte, totalSize)

	// Copy in sequence, tracking offset
	offset := 0

	// Copy topic length (2 bytes)
	copy(result[offset:], topicLength)
	offset += 2

	// Copy topic
	copy(result[offset:], topicBytes)
	offset += len(topicBytes)

	// Copy payload length (4 bytes)
	copy(result[offset:], payloadLength)
	offset += 4

	// Copy payload
	copy(result[offset:], payloadBytes)

	return result
}

type DecodedMessage struct {
	Topic   string
	Payload string
}

func DecodeMessage(c net.Conn) (DecodedMessage, error) {
	// Read topic length (2 bytes)
	topicLenBytes := make([]byte, 2)
	_, err := io.ReadFull(c, topicLenBytes)
	if err != nil {
		if err == io.EOF {
			return DecodedMessage{}, err
		}
		return DecodedMessage{}, fmt.Errorf("failed to read topic length: %w", err)
	}
	topicLen := binary.BigEndian.Uint16(topicLenBytes)

	// Read topic
	topicBytes := make([]byte, topicLen)
	_, err = io.ReadFull(c, topicBytes)
	if err != nil {
		return DecodedMessage{}, fmt.Errorf("failed to read topic: %w", err)
	}
	topic := string(topicBytes)

	// Read payload length (4 bytes)
	payloadLenBytes := make([]byte, 4)
	_, err = io.ReadFull(c, payloadLenBytes)
	if err != nil {
		return DecodedMessage{}, fmt.Errorf("failed to read payload length: %w", err)
	}
	payloadLen := binary.BigEndian.Uint32(payloadLenBytes)

	// Read payload
	payloadBytes := make([]byte, payloadLen)
	_, err = io.ReadFull(c, payloadBytes)
	if err != nil {
		return DecodedMessage{}, fmt.Errorf("failed to read payload: %w", err)
	}
	payload := string(payloadBytes)

	return DecodedMessage{
		Topic:   topic,
		Payload: payload,
	}, nil
}

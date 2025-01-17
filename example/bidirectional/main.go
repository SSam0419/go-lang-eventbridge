package eventclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Registration message structure
type RegisterMessage struct {
	Type       string `json:"type"`
	ClientPort int    `json:"client_port"`
	ClientHost string `json:"client_host"`
}

type Client struct {
	// Connection to the main server
	serverConn net.Conn
	serverAddr string

	// Client's own listener for receiving messages
	listener     net.Listener
	listenerPort int
	listenerHost string

	mu       sync.RWMutex
	handlers map[string][]MessageHandler
	shutdown chan struct{}
}

func NewClient(serverPort int) (*Client, error) {
	// First, create a listener on any available port
	listener, err := net.Listen("tcp", ":0") // ":0" means any available port
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	// Get the actual address that was assigned
	listenerAddr := listener.Addr().(*net.TCPAddr)

	client := &Client{
		serverAddr:   fmt.Sprintf(":%d", serverPort),
		listener:     listener,
		listenerPort: listenerAddr.Port,
		listenerHost: getLocalIP(), // Helper function to get local IP
		handlers:     make(map[string][]MessageHandler),
		shutdown:     make(chan struct{}),
	}

	// Connect to the main server
	if err := client.connectToServer(); err != nil {
		listener.Close()
		return nil, err
	}

	// Register with the server
	if err := client.registerWithServer(); err != nil {
		listener.Close()
		client.serverConn.Close()
		return nil, err
	}

	// Start listening for incoming messages
	go client.startListening()

	return client, nil
}

func (c *Client) connectToServer() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	c.serverConn = conn
	return nil
}

func (c *Client) registerWithServer() error {
	// Send registration message with our listener's address
	regMsg := RegisterMessage{
		Type:       "REGISTER",
		ClientPort: c.listenerPort,
		ClientHost: c.listenerHost,
	}

	encoded, err := json.Marshal(regMsg)
	if err != nil {
		return fmt.Errorf("failed to encode registration: %w", err)
	}

	_, err = c.serverConn.Write(encoded)
	return err
}

func (c *Client) startListening() {
	for {
		select {
		case <-c.shutdown:
			return
		default:
			conn, err := c.listener.Accept()
			if err != nil {
				log.Printf("Accept error: %v", err)
				continue
			}
			go c.handleIncomingMessage(conn)
		}
	}
}

type ClientInfo struct {
	ServerConn   net.Conn // Connection for receiving messages from client
	ListenerHost string   // Client's host address for sending messages
	ListenerPort int      // Client's port for sending messages
	LastSeen     time.Time
}

type Server struct {
	listener net.Listener
	clients  map[string]*ClientInfo // key: clientID
	mu       sync.RWMutex
}

func (s *Server) handleRegistration(conn net.Conn, data []byte) error {
	var regMsg RegisterMessage
	if err := json.Unmarshal(data, &regMsg); err != nil {
		return err
	}

	clientID := conn.RemoteAddr().String()

	s.mu.Lock()
	s.clients[clientID] = &ClientInfo{
		ServerConn:   conn,
		ListenerHost: regMsg.ClientHost,
		ListenerPort: regMsg.ClientPort,
		LastSeen:     time.Now(),
	}
	s.mu.Unlock()

	log.Printf("Client registered: %s (listening on %s:%d)",
		clientID, regMsg.ClientHost, regMsg.ClientPort)
	return nil
}

func (s *Server) SendToClient(clientID, message string) error {
	s.mu.RLock()
	clientInfo, exists := s.clients[clientID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found")
	}

	// Create new connection to client's listener
	clientAddr := fmt.Sprintf("%s:%d",
		clientInfo.ListenerHost,
		clientInfo.ListenerPort)

	conn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to client: %w", err)
	}
	defer conn.Close()

	// Send the message
	_, err = conn.Write([]byte(message))
	return err
}

// Helper function to get local IP
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

// Usage example
func Example() {
	// Start server
	server, _ := NewServer(8080)
	go server.Run()

	// Create client
	client, err := NewClient(8080)
	if err != nil {
		log.Fatal(err)
	}

	// The client automatically:
	// 1. Creates a listener on a random available port
	// 2. Connects to the server
	// 3. Sends its listener address to the server
	// 4. Starts listening for incoming messages

	// Server can now send messages to the client
	clientID := client.serverConn.RemoteAddr().String()
	server.SendToClient(clientID, "Hello client!")
}

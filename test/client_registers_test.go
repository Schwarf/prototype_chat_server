package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"testing"
)

type RegisterRequest struct {
	Secret   string `json:"secret"`
	Username string `json:"username"`
}

type RegisterResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Salt     string `json:"salt"`
}

type Message struct {
	ClientID    int    `json:"clientId"`
	Text        string `json:"text"`
	TimestampMs int64  `json:"timestamp_ms"`
	Hash        string `json:"hash"`
}

func registerClient(secret, username string, t *testing.T) (*RegisterResponse, error) {
	url := "http://localhost:8080/register"
	reqBody := RegisterRequest{Secret: secret, Username: username}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to register client: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	var registerResponse RegisterResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	t.Logf("Registered client: %s", registerResponse.Username)
	return &registerResponse, nil
}

func connectWebSocket(token string, t *testing.T) *websocket.Conn {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	url := "ws://localhost:8080/ws"

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket: %v", err)
	}
	return conn
}

func disconnectWebSocket(conn *websocket.Conn, t *testing.T) {
	err := conn.Close()
	if err != nil {
		t.Fatalf("failed to close WebSocket connection: %v", err)
	}
}

func TestClient(t *testing.T) {
	secret := os.Getenv("CHAT_SERVER_SECRET")
	username := os.Getenv("CHAT_SERVER_USERNAME")

	if secret == "" || username == "" {
		t.Fatalf("environment variables CHAT_SERVER_SECRET and CHAT_SERVER_USERNAME must be set")
	}
	registerResponse, err := registerClient(secret, username, t)
	if err != nil {
		t.Fatalf("failed to register client: %v", err)
	}

	// Connect to WebSocket
	conn := connectWebSocket(registerResponse.Token, t)
	t.Log("Connected to WebSocket")

	// Send a test message
	message := "Hello, Server!"

	hash := authentication.GenerateHash(message, registerResponse.Salt)
	msg := Message{
		ClientID:    registerResponse.ID,
		Text:        message,
		TimestampMs: 0,
		Hash:        hash,
	}
	t.Log("Client ID", msg.ClientID)
	msgBytes, err := json.Marshal(msg)

	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// Read acknowledgment
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	t.Logf("Received acknowledgment from server: %s\n", response)

	// Disconnect the WebSocket
	disconnectWebSocket(conn, t)
	t.Log("Disconnected from WebSocket")

	// Reconnect to WebSocket
	conn = connectWebSocket(registerResponse.Token, t)
	t.Log("Reconnected to WebSocket")

	// Send another test message
	message = "Hello again, Server!"
	hash = authentication.GenerateHash(message, registerResponse.Salt)
	msg = Message{
		ClientID:    registerResponse.ID,
		Text:        message,
		TimestampMs: 0,
		Hash:        hash,
	}
	msgBytes, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// Read acknowledgment
	_, response, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	t.Logf("Received acknowledgment from server: %s\n", response)

	// Disconnect the WebSocket again
	disconnectWebSocket(conn, t)
	t.Log("Disconnected from WebSocket")
}

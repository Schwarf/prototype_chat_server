package test

import (
	"encoding/json"
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
	"github.com/gorilla/websocket"
	"os"
	"testing"
)

func sendMessage(clientID int, conn *websocket.Conn, text, salt string, t *testing.T) {
	hash := authentication.GenerateHash(text, salt)
	msg := Message{
		ClientID:    clientID,
		Text:        text,
		TimestampMs: 0, // Use actual timestamp if needed
		Hash:        hash,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
}

func readMessage(conn *websocket.Conn, t *testing.T) Message {
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}
	return msg
}

func TestTwoClientsMessageExchange(t *testing.T) {
	// Set up environment
	secret := os.Getenv("CHAT_SERVER_SECRET")
	if secret == "" {
		t.Fatalf("environment variable CHAT_SERVER_SECRET must be set")
	}

	// Register Client A
	usernameA := "ClientA"
	registerResponseA, err := registerClient(secret, usernameA, t)
	if err != nil {
		t.Fatalf("failed to register Client A: %v", err)
	}
	connA := connectWebSocket(registerResponseA.Token, t)
	defer disconnectWebSocket(connA, t)
	t.Log("Client A connected to WebSocket")

	// Register Client B
	usernameB := "ClientB"
	registerResponseB, err := registerClient(secret, usernameB, t)
	if err != nil {
		t.Fatalf("failed to register Client B: %v", err)
	}
	connB := connectWebSocket(registerResponseB.Token, t)
	defer disconnectWebSocket(connB, t)
	t.Log("Client B connected to WebSocket")

	// Client A sends a message
	sendMessage(registerResponseA.ID, connA, "Hello from Client A", registerResponseA.Salt, t)
	t.Log("Client A sent a message")

	// Client B reads the message
	msg := readMessage(connB, t)
	t.Logf("Client B received message: %s", msg.Text)

	// Check the received message
	if msg.Text != "Hello from Client A" {
		t.Fatalf("Client B received incorrect message: %s", msg.Text)
	}

	// Client B sends a message
	sendMessage(registerResponseB.ID, connB, "Hello from Client B", registerResponseB.Salt, t)
	t.Log("Client B sent a message")

	// Client A reads the message
	msg = readMessage(connA, t)
	t.Logf("Client A received message: %s", msg.Text)

	// Check the received message
	if msg.Text != "Hello from Client B" {
		t.Fatalf("Client A received incorrect message: %s", msg.Text)
	}
}

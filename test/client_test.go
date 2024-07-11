package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"testing"
)

type RegisterRequest struct {
	Secret   string `json:"secret"`
	Username string `json:"username"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

func registerClient(secret, username string) (string, error) {
	url := "http://localhost:8080/register"
	reqBody := RegisterRequest{Secret: secret, Username: username}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("Test: Failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", fmt.Errorf("Test: Failed to register client: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	var registerResponse RegisterResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResponse)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return registerResponse.Token, nil
}

func connectWebSocket(token string) {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	url := "ws://localhost:8080/ws"

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Send a test message
	message := "Hello, Server!"
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Read response
	_, response, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("Failed to read message: %v", err)
	}

	fmt.Printf("Received message from server: %s\n", response)
}

func TestClient(t *testing.T) {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <secret> <username>", os.Args[0])
	}

	secret := os.Args[1]
	username := os.Args[2]

	token, err := registerClient(secret, username)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}

	connectWebSocket(token)
}

package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

type PresenceResponse struct {
	Status string `json:"status"`
}

func checkPresence(clientID int, t *testing.T) (*PresenceResponse, error) {
	url := fmt.Sprintf("http://localhost:8080/check_presence?client_id=%d", clientID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to check presence: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("presence check failed with status code: %d", resp.StatusCode)
	}

	var presenceResponse PresenceResponse
	err = json.NewDecoder(resp.Body).Decode(&presenceResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode presence response: %v", err)
	}

	t.Logf("Presence check for client %d: %s", clientID, presenceResponse.Status)
	return &presenceResponse, nil
}

func TestCheckPresence(t *testing.T) {
	secret := os.Getenv("CHAT_SERVER_SECRET")
	username := os.Getenv("CHAT_SERVER_USERNAME")

	if secret == "" || username == "" {
		t.Fatalf("environment variables CHAT_SERVER_SECRET and CHAT_SERVER_USERNAME must be set")
	}

	// Register a client
	registerResponse, err := registerClient(secret, username, t)
	if err != nil {
		t.Fatalf("failed to register client: %v", err)
	}

	clientID := registerResponse.ID

	// Check presence (should be not present initially)
	presenceResponse, err := checkPresence(clientID, t)
	if err != nil {
		t.Fatalf("failed to check presence: %v", err)
	}
	if presenceResponse.Status != "not_present" {
		t.Fatalf("unexpected presence status: %v", presenceResponse.Status)
	}

	// Connect to WebSocket
	conn := connectWebSocket(registerResponse.Token, t)
	t.Log("Connected to WebSocket")

	// Check presence (should be present after connection)
	presenceResponse, err = checkPresence(clientID, t)
	if err != nil {
		t.Fatalf("failed to check presence: %v", err)
	}
	if presenceResponse.Status != "present" {
		t.Fatalf("unexpected presence status: %v", presenceResponse.Status)
	}

	// Disconnect the WebSocket
	disconnectWebSocket(conn, t)
	t.Log("Disconnected from WebSocket")

	// Check presence again (should be not present after disconnection)
	presenceResponse, err = checkPresence(clientID, t)
	if err != nil {
		t.Fatalf("failed to check presence: %v", err)
	}
	if presenceResponse.Status != "not_present" {
		t.Fatalf("unexpected presence status: %v", presenceResponse.Status)
	}
}

package handlers

import (
	"encoding/json"
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/google/uuid"
	"log"
	"net/http"
)

func RegisterClientHandler(writer http.ResponseWriter, request *http.Request) {
	// Expected request send to endpoint
	var submittedRequest struct {
		Secret   string `json:"secret"`
		Username string `json:"username"`
	}
	err := json.NewDecoder(request.Body).Decode(&submittedRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the submitted secret
	if !authentication.IsSecretValid(submittedRequest.Secret) {
		http.Error(writer, "Invalid secret", http.StatusUnauthorized)
		return
	}

	if len(submittedRequest.Username) < 6 || !authentication.IsAlphaNumeric(submittedRequest.Username) {
		http.Error(writer, "Invalid username", http.StatusUnauthorized)
		return
	}

	token, err := authentication.GenerateToken(submittedRequest.Username)
	if err != nil {
		http.Error(writer, "Error generating token", http.StatusInternalServerError)
		return
	}
	clientID := uuid.New().String()
	client := models.Client{
		Username: submittedRequest.Username,
		Token:    token,
	}
	authentication.RegisterClient(clientID, client)
	authentication.RemoveSecret(submittedRequest.Secret)

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(client)
	log.Println("Client is registered")
}

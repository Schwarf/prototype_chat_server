package handlers

import (
	"encoding/json"
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/Schwarf/prototype_chat_server/internal/storage"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type RegisterHandler struct {
	database    *storage.DB
	HandlerFunc func(db *storage.DB, w http.ResponseWriter, r *http.Request)
}

func (handler RegisterHandler) RegisterClient(w http.ResponseWriter, r *http.Request) {
	handler.HandlerFunc(handler.database, w, r)
}

func RegisterClient(database *storage.DB, writer http.ResponseWriter, request *http.Request) {
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

	salt := uuid.New().String()
	clientID, err := storage.AddClient(database, submittedRequest.Username, token, salt)

	if err != nil {
		http.Error(writer, "Error adding client to database", http.StatusInternalServerError)
		return
	}

	client := models.Client{
		ID:       clientID,
		Username: submittedRequest.Username,
		Token:    token,
		Salt:     salt,
	}

	authentication.RegisterClient(clientID, client)
	authentication.RemoveSecret(submittedRequest.Secret)
	log.Println("Bad request8")
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(client)
	log.Println("Client has been registered successfully")
}

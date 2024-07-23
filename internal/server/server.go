package server

import (
	"encoding/json"
	"fmt"
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
	"github.com/Schwarf/prototype_chat_server/internal/handlers"
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/Schwarf/prototype_chat_server/internal/storage"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Server struct {
	config    *config.ServerConfig
	clients   map[*models.ChatClient]bool
	broadcast chan models.Message
	mutex     sync.Mutex
	database  *storage.DB
	upgrader  websocket.Upgrader
}

func NewServer(serverConfig *config.ServerConfig, dataBase *storage.DB) *Server {
	return &Server{
		config:    serverConfig,
		clients:   make(map[*models.ChatClient]bool),
		broadcast: make(chan models.Message),
		database:  dataBase,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}
func (s *Server) homepage(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Welcome to Schwarf's WebSocket chat server!")
}

func (s *Server) broadcastMessage(message models.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for client := range s.clients {
		if client.Online {
			msgJSON, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message to JSON: %v", err)
				continue
			}
			if err := client.SendMessage(websocket.TextMessage, msgJSON); err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				client.Online = false
			}
		}
	}
}

func (s *Server) retryUndeliveredMessages() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	undeliveredMessages, err := storage.RetrieveUndeliveredMessages(s.database)
	if err != nil {
		log.Printf("Failed to retrieve undelivered messages: %v", err)
		return
	}

	for _, message := range undeliveredMessages {
		for client := range s.clients {
			if client.Online {
				msgJSON, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshaling message to JSON: %v", err)
					continue
				}
				if err := client.SendMessage(websocket.TextMessage, msgJSON); err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					client.Online = false
				} else {
					storage.UpdateMessageStatus(s.database, message.DBID, true)
				}
			}
		}
	}
}

func (s *Server) handleMessages() {
	for {
		select {
		case message := <-s.broadcast:
			s.broadcastMessage(message)
		case <-time.After(time.Second * 3):
			s.retryUndeliveredMessages()
		}
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.homepage)
	http.HandleFunc("/check_presence", func(writer http.ResponseWriter, request *http.Request) {
		handlers.CheckPresence(s.clients, &s.mutex, writer, request)
	})
	http.HandleFunc("/register", func(writer http.ResponseWriter, request *http.Request) {
		handlers.RegisterClient(s.database, writer, request)
	})
	http.HandleFunc("/ws", s.websocketEndpoint)
	log.Println("Starting server on port", s.config.Port)
	go s.handleMessages()
	return http.ListenAndServe(s.config.Port, nil)
}

func (s *Server) Stop() error {
	fmt.Println("Stopping server...")

	// Drop all tables in the database
	//if err := storage.DropAllTables(s.database.DB); err != nil {
	//	log.Printf("Failed to drop tables: %v", err)
	//	return err
	//}

	// Close database connection
	s.database.Close()

	// Additional cleanup tasks if needed
	fmt.Println("Server stopped.")
	return nil
}

func (s *Server) storeMessage(message models.Message) error {
	if err := storage.StoreMessage(s.database, message); err != nil {
		log.Printf("Storing message failed! Error: %v", err)
		return err
	}
	return nil
}

func (s *Server) websocketEndpoint(writer http.ResponseWriter, request *http.Request) {
	connection, err := s.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer connection.Close()
	authenticationHeader := request.Header.Get("Authorization")
	if authenticationHeader == "" {
		http.Error(writer, "Authorization header is missing", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authenticationHeader, "Bearer ")

	if token == "" {
		log.Println("Missing token")
		http.Error(writer, "Missing token", http.StatusUnauthorized)
		return
	}

	clientID, salt, err := storage.GetClientIDAndSalt(s.database, token)
	if err != nil {
		log.Printf("Failed to get client ID by token: %v, %s", err, token)
		http.Error(writer, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if client with this ID is already connected.
	s.mutex.Lock()
	for client := range s.clients {
		if client.ID == clientID {
			s.mutex.Unlock()
			log.Printf("Client %d is already connected. Declining new connection attempt.", clientID)
			connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client already connected"))
			return
		}
	}
	client := &models.ChatClient{ID: clientID, Connection: connection}
	s.clients[client] = true
	s.mutex.Unlock()

	log.Printf("ChatClient %d connected", client.ID)
	defer func() {
		log.Printf("ChatClient %d disconnected", client.ID)
		s.mutex.Lock()
		delete(s.clients, client)
		s.mutex.Unlock()
	}()

	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		expectedHash := authentication.GenerateHash(msg.Text, salt)
		if msg.Hash != expectedHash {
			log.Printf("Invalid hash for message from client %s", clientID)
			continue
		}

		log.Printf("Received message from client %d at %s: %s\n", clientID, time.Now().Format(time.RFC3339), message)
		s.broadcast <- msg
		if err := s.storeMessage(msg); err != nil {
			log.Printf("Failed to store message! Error: %v", err)
		}
		ack := fmt.Sprintf("Message from client %d received at %s", client.ID, time.Now().Format(time.RFC3339))
		if err := client.SendMessage(websocket.TextMessage, []byte(ack)); err != nil {
			log.Printf("Error sending acknowledgment to WebSocket: %v", err)
		}
	}
}

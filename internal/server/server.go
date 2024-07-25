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
	"os"
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

func (server *Server) homepage(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Welcome to Schwarf'server WebSocket chat server!")
}

func (server *Server) broadcastMessage(message models.Message) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for client := range server.clients {
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

func (server *Server) retryUndeliveredMessages() {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	undeliveredMessages, err := storage.RetrieveUndeliveredMessages(server.database)
	if err != nil {
		log.Printf("Failed to retrieve undelivered messages: %v", err)
		return
	}

	for _, message := range undeliveredMessages {
		for client := range server.clients {
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
					storage.UpdateMessageStatus(server.database, message.DBID, true)
				}
			}
		}
	}
}

func (server *Server) handleMessages() {
	for {
		select {
		case message := <-server.broadcast:
			server.broadcastMessage(message)
		case <-time.After(time.Second * 3):
			server.retryUndeliveredMessages()
		}
	}
}

func (server *Server) Start() error {
	http.HandleFunc("/", server.homepage)
	http.HandleFunc("/check_presence", func(writer http.ResponseWriter, request *http.Request) {
		handlers.CheckPresence(server.clients, &server.mutex, writer, request)
	})
	http.HandleFunc("/register", func(writer http.ResponseWriter, request *http.Request) {
		handlers.RegisterClient(server.database, writer, request)
	})
	http.HandleFunc("/ws", server.websocketEndpoint)
	log.Println("Starting server on port", server.config.Port)
	go server.handleMessages()
	return http.ListenAndServe(server.config.Port, nil)
}

func (server *Server) Stop() error {
	fmt.Println("Stopping server...")

	envVariable := os.Getenv("APP_ENV")
	if envVariable != "" {
		//Drop all tables in the database
		if err := storage.DropAllTables(server.database.DB); err != nil {
			log.Printf("Failed to drop tables: %v", err)
			return err
		}
	}

	// Close database connection
	server.database.Close()

	// Additional cleanup tasks if needed
	fmt.Println("Server stopped.")
	return nil
}

func (server *Server) storeMessage(message models.Message) error {
	if err := storage.StoreMessage(server.database, message); err != nil {
		log.Printf("Storing message failed! Error: %v", err)
		return err
	}
	return nil
}

func (server *Server) authenticateClient(request *http.Request, writer http.ResponseWriter) (int, string, error) {
	authenticationHeader := request.Header.Get("Authorization")
	if authenticationHeader == "" {
		http.Error(writer, "Authorization header is missing", http.StatusUnauthorized)
		return 0, "", fmt.Errorf("authorization header is missing")
	}

	token := strings.TrimPrefix(authenticationHeader, "Bearer ")
	if token == "" {
		log.Println("Missing token")
		http.Error(writer, "Missing token", http.StatusUnauthorized)
		return 0, "", fmt.Errorf("missing token")
	}
	clientID, salt, err := storage.GetClientIDAndSalt(server.database, token)
	if err != nil {
		log.Printf("Failed to get clientID by token: %v, %s", err, token)
		http.Error(writer, "Invalid token", http.StatusUnauthorized)
		return 0, "", fmt.Errorf("invalid token")
	}
	return clientID, salt, nil
}

func (server *Server) isClientAlreadyConnected(clientID int, connection *websocket.Conn) bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for client := range server.clients {
		if client.ID == clientID {
			log.Printf("Client %d is already connected. Declining new connection attempt.", clientID)
			err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client already connected"))
			if err != nil {
				log.Printf("Failed to notify client %d about existing connection: %v", clientID, err)
			}
			return true
		}
	}
	return false
}

func (server *Server) addChatClient(connection *websocket.Conn, clientID int) *models.ChatClient {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	chatClient := &models.ChatClient{ID: clientID, Connection: connection, Online: true}
	server.clients[chatClient] = true
	log.Printf("Added connection for new ChatClient %d", clientID)
	return chatClient
}

func (server *Server) removeChatClient(chatClient *models.ChatClient) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.clients, chatClient)
	log.Printf("ChatClient %d disconnected", chatClient.ID)
}

func (server *Server) readMessages(chatClient *models.ChatClient, salt string) {
	for {
		_, message, err := chatClient.Connection.ReadMessage()
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
			log.Printf("Invalid hash for message from chatClient %server", chatClient.ID)
			continue
		}

		log.Printf("Received message from chatClient %d at %server: %server\n", chatClient.ID, time.Now().Format(time.RFC3339), message)
		server.broadcast <- msg
		if err := server.storeMessage(msg); err != nil {
			log.Printf("Failed to store message! Error: %v", err)
		}
		ack := fmt.Sprintf("Message from chatClient %d received at %server", chatClient.ID, time.Now().Format(time.RFC3339))
		if err := chatClient.SendMessage(websocket.TextMessage, []byte(ack)); err != nil {
			log.Printf("Error sending acknowledgment to WebSocket: %v", err)
		}
	}

}

func (server *Server) websocketEndpoint(writer http.ResponseWriter, request *http.Request) {
	connection, err := server.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer connection.Close()
	clientID, salt, err := server.authenticateClient(request, writer)
	if err != nil {
		log.Printf("Failed to authenticate chatClient: %v", err)
		return
	}

	if server.isClientAlreadyConnected(clientID, connection) {
		return
	}

	chatClient := server.addChatClient(connection, clientID)
	defer server.removeChatClient(chatClient)

	server.readMessages(chatClient, salt)
}

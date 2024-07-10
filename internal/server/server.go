package server

import (
	"fmt"
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/Schwarf/prototype_chat_server/internal/storage"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	config    *config.ServerConfig
	clients   map[*Client]bool
	broadcast chan models.Message
	mutex     sync.Mutex
	database  *storage.DB
	upgrader  websocket.Upgrader
}

func NewServer(serverConfig *config.ServerConfig, dataBase *storage.DB) *Server {
	return &Server{
		config:    serverConfig,
		clients:   make(map[*Client]bool),
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

func (s *Server) broadcastMessage(messageType int, message []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for client := range s.clients {
		if err := client.Connection.WriteMessage(messageType, message); err != nil {
			log.Printf("Error writing to WebSocket: %v", err)
			client.Connection.Close()
			delete(s.clients, client)
		}
	}
}
func (s *Server) handleMessages() {
	for {
		msg := <-s.broadcast
		s.broadcastMessage(websocket.TextMessage, []byte(msg.Text))
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.homepage)
	http.HandleFunc("/ws", s.websocketEndpoint)
	log.Println("Starting server on port", s.config.Port)
	go s.handleMessages()
	return http.ListenAndServe(s.config.Port, nil)
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
	clientID := fmt.Sprintf("Client-%d", time.Now().UnixNano())
	client := &Client{ID: clientID, Connection: connection, Server: s}

	s.mutex.Lock()
	s.clients[client] = true
	s.mutex.Unlock()

	log.Printf("Client %s connected", client.ID)
	defer func() {
		log.Printf("Client %s disconnected", client.ID)
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
		timestamp := time.Now().Unix()
		log.Printf("Received message from client %s at %s: %s\n", client.ID, time.Now().Format(time.RFC3339), message)
		msg := models.Message{ChatID: clientID, Sender: client.ID, Text: string(message), Timestamp: timestamp, Hash: "somehash"}
		s.broadcast <- msg
		if err := s.storeMessage(msg); err != nil {
			log.Printf("Failed to store message! Error: %v", err)
		}
	}
}

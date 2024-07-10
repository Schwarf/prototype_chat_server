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

func (s *Server) Start() error {
	http.HandleFunc("/", s.homepage)
	http.HandleFunc("/ws", s.websocketEndpoint)
	log.Println("Starting server on port", s.config.Port)
	go s.handleMessages()
	return http.ListenAndServe(s.config.Port, nil)
}

func (s *Server) websocketEndpoint(writer http.ResponseWriter, request *http.Request) {
	connections, err := s.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer connections.Close()
}

func (s *Server) handleMessages() {

}

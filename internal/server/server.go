package server

import (
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/Schwarf/prototype_chat_server/internal/storage"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
	"github.com/gorilla/websocket"
	"sync"
)

type Server struct {
	config    *config.Config
	clients   map[*Client]bool
	broadcast chan models.Message
	mu        sync.Mutex
	db        *storage.DB
	upgrader  websocket.Upgrader
}

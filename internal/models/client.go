package models

import (
	"github.com/Schwarf/prototype_chat_server/internal/server"
	"github.com/gorilla/websocket"
)

type ChatClient struct {
	ID         string
	Connection *websocket.Conn
	Server     *server.Server
	Online     bool
}

type Client struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (c *ChatClient) SendMessage(messageType int, message []byte) error {
	return c.Connection.WriteMessage(messageType, message)
}

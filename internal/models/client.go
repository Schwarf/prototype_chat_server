package models

import (
	"github.com/gorilla/websocket"
)

type ChatClient struct {
	ID         string
	Connection *websocket.Conn
	Online     bool
}

type Client struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (c *ChatClient) SendMessage(messageType int, message []byte) error {
	return c.Connection.WriteMessage(messageType, message)
}

package models

import (
	"github.com/gorilla/websocket"
)

type ChatClient struct {
	ID         int
	Connection *websocket.Conn
	Online     bool
}

type Client struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (c *ChatClient) SendMessage(messageType int, message []byte) error {
	return c.Connection.WriteMessage(messageType, message)
}

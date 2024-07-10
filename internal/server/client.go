package server

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	ID         string
	Connection *websocket.Conn
	Server     *Server
	Online     bool
}

func (c *Client) SendMessage(messageType int, message []byte) error {
	return c.Connection.WriteMessage(messageType, message)
}

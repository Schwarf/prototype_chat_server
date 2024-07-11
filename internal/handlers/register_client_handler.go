package handlers

import (
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"sync"
)

var clients = make(map[*models.ChatClient]bool)
var mu sync.Mutex
var broadcast = make(chan models.Message)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

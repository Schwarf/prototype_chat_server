package handlers

import (
	"fmt"
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"net/http"
	"sync"
)

type CheckPresenceHandler struct {
	clients     map[*models.ChatClient]bool
	clientsMu   *sync.Mutex
	HandlerFunc func(clients map[*models.ChatClient]bool, clientsMu *sync.Mutex, w http.ResponseWriter, r *http.Request)
}

func (handler CheckPresenceHandler) CheckPresence(w http.ResponseWriter, r *http.Request) {
	handler.HandlerFunc(handler.clients, handler.clientsMu, w, r)
}

func CheckPresence(clients map[*models.ChatClient]bool, clientsMutex *sync.Mutex, w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		if fmt.Sprintf("%d", client.ID) == clientID && client.Online {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("present"))
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not_present"))
}

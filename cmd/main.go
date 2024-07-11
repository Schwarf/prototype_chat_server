package main

import (
	"github.com/Schwarf/prototype_chat_server/internal/server"
	"github.com/Schwarf/prototype_chat_server/internal/storage"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
	"log"
)

func main() {
	// Load server configuration
	serverConfig := config.LoadServerConfig()
	databaseConfig, err := config.LoadDataBaseConfig()
	if err != nil {
		log.Fatalf("Database config could not be loaded")
		return
	}

	// Initialize the database
	db, err := storage.ConnectToDatabase(databaseConfig)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Create and start the server
	srv := server.NewServer(serverConfig, db)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

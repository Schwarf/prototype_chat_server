package main

import (
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
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
	}

	// Initialize the database
	db, err := storage.ConnectToDatabase(databaseConfig)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Creates MessageTable if is not existing yet
	err = storage.CreateMessagesTable(db)
	if err != nil {
		log.Fatalf("CreatingMessageTable failed: %v", err)
	}

	authentication.LoadSecrets()

	// Create and start the server
	srv := server.NewServer(serverConfig, db)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

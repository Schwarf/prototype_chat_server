package main

import (
	"fmt"
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
	"github.com/Schwarf/prototype_chat_server/internal/server"
	"github.com/Schwarf/prototype_chat_server/internal/storage"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	authentication.LoadSecrets()

	// Create and start the server
	srv := server.NewServer(serverConfig, db)
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	// Signal received, stop the server
	fmt.Println("\nReceived termination signal. Stopping server...")
	if err := srv.Stop(); err != nil {
		fmt.Printf("Error stopping server: %v\n", err)
	}

	fmt.Println("Server stopped gracefully.")
}

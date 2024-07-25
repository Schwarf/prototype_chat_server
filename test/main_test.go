package test

import (
	"github.com/Schwarf/prototype_chat_server/internal/authentication"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Schwarf/prototype_chat_server/internal/server"
	"github.com/Schwarf/prototype_chat_server/internal/storage"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
)

var srv *server.Server
var once sync.Once

func setup() {
	once.Do(func() {
		// Load server configuration
		os.Setenv("APP_ENV", "test")
		authentication.LoadSecrets()
		serverConfig := config.LoadServerConfig()
		databaseConfig, err := config.LoadDataBaseConfig()
		if err != nil {
			log.Fatalf("Database config could not be loaded: %v", err)
		}

		// Initialize the database
		db, err := storage.ConnectToDatabase(databaseConfig)
		if err != nil {
			log.Fatalf("Database connection failed: %v", err)
		}

		// Create and start the server
		srv = server.NewServer(serverConfig, db)
		go func() {
			if err := srv.Start(); err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		}()
		log.Println("Started server in test!!!")
		log.Println("Sleep for 0.5 seconds to get server up!!!")
		time.Sleep(500 * time.Millisecond)
	})
}

func teardown() {
	if srv != nil {
		if err := srv.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
	}
}

func TestMain(m *testing.M) {
	// Setup
	setup()

	// Run tests
	code := m.Run()

	// Teardown
	teardown()

	// Exit with the appropriate code
	os.Exit(code)
}

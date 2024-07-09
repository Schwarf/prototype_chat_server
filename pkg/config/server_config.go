package config

import (
	"os"
)

type ServerConfig struct {
	Port string
}

func LoadServerConfig() *ServerConfig {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	return &ServerConfig{Port: port}
}

package config

import (
	"encoding/json"
	"log"
	"os"
)

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

func LoadDataBaseConfig() (*DatabaseConfig, error) {

	envVariable := os.Getenv("APP_ENV")
	var filePath string
	switch envVariable {
	case "test":
		filePath = "/home/andreas/Documents/database_access/postgres_test_config.json"
	default:
		filePath = "/home/andreas/Documents/database_access/postgres_config.json"
	}

	var config DatabaseConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening config file: %v", err)
		return nil, err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		log.Printf("Error parsing config file: %v", err)
		return nil, err
	}
	return &config, err
}

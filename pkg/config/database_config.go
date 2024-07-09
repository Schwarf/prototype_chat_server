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
	filePath := "/home/andreas/Documents/database_access/postgres_config.json"
	var config *DatabaseConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening config file: %v", err)
		return config, err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(config)
	if err != nil {
		log.Printf("Error parsing config file: %v", err)
		return config, err
	}
	return config, err

}

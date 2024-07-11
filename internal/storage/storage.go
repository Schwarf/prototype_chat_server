package storage

import (
	"database/sql"
	"fmt"
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func ConnectToDatabase(config *config.DatabaseConfig) (*DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.DBName)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func CreateMessagesTable(db *DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chat_id TEXT NOT NULL,
        sender TEXT NOT NULL,
        text TEXT NOT NULL,
        timestamp_ms INTEGER NOT NULL,
        hash TEXT NOT NULL,
        delivered_to_client BOOLEAN NOT NULL DEFAULT FALSE,
    );`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func StoreMessage(db *DB, message models.Message) error {
	_, err := db.Exec("INSERT INTO messages (chat_id, sender, text, timestamp_ms, hash) VALUES (?, ?, ?, ?, ?)",
		message.ChatID, message.Sender, message.Text, message.Timestamp_ms, message.Hash)
	if err != nil {
		return err
	}
	return nil
}

func RetrieveUndeliveredMessages(db *DB) ([]models.Message, error) {
	rows, err := db.Query("SELECT id, chat_id, sender, text, timestamp_ms, hash FROM messages WHERE delivered_to_client = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.Sender, &message.Text, &message.Timestamp_ms, &message.Hash); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func UpdateMessageStatus(db *DB, messageID int, deliveredToClient bool) error {
	_, err := db.Exec("UPDATE messages SET delivered_to_client = ? WHERE id = ?", deliveredToClient, messageID)
	if err != nil {
		return err
	}
	return nil
}

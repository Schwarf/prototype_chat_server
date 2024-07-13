package storage

import (
	"database/sql"
	"fmt"
	"github.com/Schwarf/prototype_chat_server/internal/models"
	"github.com/Schwarf/prototype_chat_server/pkg/config"
	_ "github.com/lib/pq"
	"log"
)

type DB struct {
	*sql.DB
}

func ConnectToDatabase(config *config.DatabaseConfig) (*DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.DBName)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Connection to database failed: %v\n", err)
		return nil, err
	}
	if err := createTables(db); err != nil {
		log.Printf("Creating tables failed: %v\n", err)
	}

	return &DB{db}, nil
}

func createTables(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS clients (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		token TEXT UNIQUE NOT NULL,
	    salt TEXT UNIQUE NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	query = `CREATE TABLE IF NOT EXISTS chats (
		id SERIAL PRIMARY KEY,
		client_id INT REFERENCES clients(id) ON DELETE CASCADE,
		chat_id TEXT UNIQUE NOT NULL
	);`
	_, err = db.Exec(query)
	if err != nil {
		return err
	}
	query = `CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		chat_id TEXT REFERENCES chats(chat_id) ON DELETE CASCADE,
		client_id INT REFERENCES clients(id) ON DELETE CASCADE,
		text TEXT,
		timestamp_ms BIGINT,
		hash TEXT,
		delivered BOOLEAN DEFAULT FALSE
	);`
	_, err = db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func StoreMessage(db *DB, message models.Message) error {
	log.Println("Message: ", message.ChatID, message.Text)
	_, err := db.Exec("INSERT INTO messages (client_id, chat_id, text, timestamp_ms, hash) VALUES ($1, $2, $3, $4, $5)",
		message.ClientID, message.ChatID, message.Text, message.Timestamp_ms, message.Hash)
	if err != nil {
		return err
	}
	return nil
}

func AddClient(db *DB, username, token string, salt string) (int, error) {
	query := `
	INSERT INTO clients (username, token, salt)
	VALUES ($1, $2, $3)
	RETURNING id;`
	var clientID int
	err := db.QueryRow(query, username, token, salt).Scan(&clientID)
	if err != nil {
		return 0, fmt.Errorf("failed to add client: %w", err)
	}
	return clientID, nil
}

func AddChat(db *DB, clientID int, chatID string) error {
	query := `
	INSERT INTO chats (client_id, chat_id)
	VALUES ($1, $2)
	ON CONFLICT (chat_id) DO NOTHING;`
	_, err := db.Exec(query, clientID, chatID)
	if err != nil {
		return fmt.Errorf("failed to store chat: %w", err)
	}
	return nil
}

func GetClientIDAndSalt(db *DB, token string) (int, string, error) {
	var clientID int
	var salt string
	query := `
	SELECT id, salt
	FROM clients
	WHERE token = $1;`
	err := db.QueryRow(query, token).Scan(&clientID, &salt)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get client by token: %w", err)
	}
	return clientID, salt, nil
}

func RetrieveUndeliveredMessages(db *DB) ([]models.DBMessage, error) {
	rows, err := db.Query("SELECT id, chat_id, client_id, text, timestamp_ms, hash FROM messages WHERE delivered = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.DBMessage
	for rows.Next() {
		var message models.DBMessage
		if err := rows.Scan(&message.DBID, &message.ClientID, &message.ChatID, &message.Text, &message.Timestamp_ms, &message.Hash); err != nil {
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

func DropAllTables(db *sql.DB) error {
	// Query to get all table names
	query := `
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = 'public'
        AND table_type = 'BASE TABLE';
    `
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Loop through results and drop each table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		dropStmt := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", tableName)
		if _, err := db.Exec(dropStmt); err != nil {
			return err
		}
		fmt.Printf("Dropped table %s\n", tableName)
	}

	return nil
}

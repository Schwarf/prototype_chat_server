package models

type Message struct {
	ClientID     int    `json:"clientId"`
	ChatID       string `json:"chatId"`
	Text         string `json:"text"`
	Timestamp_ms int64  `json:"timestamp_ms"`
	Hash         string `json:"hash"`
}

type DBMessage struct {
	DBID         int    `json:"DBId"`
	ClientID     int    `json:"id"`
	ChatID       string `json:"chatId"`
	Text         string `json:"text"`
	Timestamp_ms int64  `json:"timestamp"`
	Hash         string `json:"hash"`
}

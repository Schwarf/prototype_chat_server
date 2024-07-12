package models

type Message struct {
	ClientID     int    `json:"id"`
	ChatID       string `json:"chatId"`
	Text         string `json:"text"`
	Timestamp_ms int64  `json:"timestamp"`
	Hash         string `json:"hash"`
}

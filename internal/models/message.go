package models

type Message struct {
	ID        int    `json:"id"`
	ChatID    string `json:"chatId"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	Hash      string `json:"hash"`
}

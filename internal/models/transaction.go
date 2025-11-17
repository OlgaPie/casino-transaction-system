package models

import "time"

type Transaction struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"user_id"`
	TransactionType string    `json:"transaction_type"` // "bet" or "win"
	Amount          float64   `json:"amount"`
	Timestamp       time.Time `json:"timestamp"`
}

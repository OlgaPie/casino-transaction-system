package models

import "time"

type TransactionType string

const (
	TransactionTypeBet TransactionType = "bet"
	TransactionTypeWin TransactionType = "win"
)

type Transaction struct {
	ID              int64           `json:"id" db:"id"`
	TransactionID   string          `json:"transaction_id" db:"transaction_id"`
	UserID          string          `json:"user_id" db:"user_id"`
	TransactionType TransactionType `json:"transaction_type" db:"transaction_type"`
	Amount          int64           `json:"amount" db:"amount"`
	Timestamp       time.Time       `json:"timestamp" db:"timestamp"`
}

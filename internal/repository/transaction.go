package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/OlgaPie/casino-transaction-system/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository interface {
	SaveTransaction(ctx context.Context, tx models.Transaction) error
	GetTransactionsByUserID(ctx context.Context, userID string, txType string) ([]models.Transaction, error)
	GetAllTransactions(ctx context.Context, txType string) ([]models.Transaction, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) TransactionRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) SaveTransaction(ctx context.Context, tx models.Transaction) error {
	sql := `INSERT INTO transactions (user_id, transaction_type, amount, "timestamp") VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(ctx, sql, tx.UserID, tx.TransactionType, tx.Amount, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("could not save transaction: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetTransactionsByUserID(ctx context.Context, userID string, txType string) ([]models.Transaction, error) {
	baseSQL := `SELECT id, user_id, transaction_type, amount, "timestamp" FROM transactions WHERE user_id = $1`
	args := []interface{}{userID}

	if txType != "" {
		baseSQL += " AND transaction_type = $2"
		args = append(args, txType)
	}

	rows, err := r.db.Query(ctx, baseSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.TransactionType, &tx.Amount, &tx.Timestamp); err != nil {
			log.Printf("could not scan row: %v", err)
			continue
		}
		transactions = append(transactions, tx)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", rows.Err())
	}

	return transactions, nil
}

func (r *postgresRepository) GetAllTransactions(ctx context.Context, txType string) ([]models.Transaction, error) {
	baseSQL := `SELECT id, user_id, transaction_type, amount, "timestamp" FROM transactions`
	var args []interface{}

	if txType != "" {
		baseSQL += " WHERE transaction_type = $1"
		args = append(args, txType)
	}

	baseSQL += ` ORDER BY "timestamp" DESC`

	rows, err := r.db.Query(ctx, baseSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query all transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.TransactionType, &tx.Amount, &tx.Timestamp); err != nil {
			log.Printf("could not scan row: %v", err)
			continue
		}
		transactions = append(transactions, tx)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", rows.Err())
	}

	return transactions, nil
}

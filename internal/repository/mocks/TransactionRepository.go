package mocks

import (
	"context"

	"github.com/OlgaPie/casino-transaction-system/internal/models"
	"github.com/stretchr/testify/mock"
)

type TransactionRepository struct {
	mock.Mock
}

func (m *TransactionRepository) SaveTransaction(ctx context.Context, tx models.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *TransactionRepository) GetTransactionsByUserID(ctx context.Context, userID string, txType string) ([]models.Transaction, error) {
	args := m.Called(ctx, userID, txType)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *TransactionRepository) GetAllTransactions(ctx context.Context, txType string) ([]models.Transaction, error) {
	args := m.Called(ctx, txType)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

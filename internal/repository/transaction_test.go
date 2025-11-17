package repository

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/OlgaPie/casino-transaction-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB запускает контейнер с PostgreSQL для тестов.
func setupTestDB(ctx context.Context) (*pgxpool.Pool, func()) {
	// Создаем запрос на запуск контейнера Postgres
	pgContainer, err := postgres.Run(ctx, "postgres:14-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Получаем connection string для подключения к тестовой базе
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	// Создаем пул соединений
	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect to test db: %s", err)
	}

	// Создаем таблицу
	_, err = dbpool.Exec(ctx, `CREATE TABLE transactions (
        id SERIAL PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        transaction_type VARCHAR(10) NOT NULL,
        amount NUMERIC(15, 2) NOT NULL,
        "timestamp" TIMESTAMPTZ NOT NULL
    );`)
	if err != nil {
		log.Fatalf("failed to create table: %s", err)
	}

	// Возвращаем пул соединений и функцию для очистки (остановки контейнера)
	cleanup := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}

	return dbpool, cleanup
}

func TestPostgresRepository(t *testing.T) {
	ctx := context.Background()
	dbpool, cleanup := setupTestDB(ctx)
	defer cleanup()

	repo := NewPostgresRepository(dbpool)

	// --- Тестируем SaveTransaction ---
	t.Run("should save and retrieve a transaction", func(t *testing.T) {
		tx := models.Transaction{
			UserID:          "user123",
			TransactionType: "bet",
			Amount:          100.50,
			Timestamp:       time.Now(),
		}

		err := repo.SaveTransaction(ctx, tx)
		require.NoError(t, err) // require прерывает тест при ошибке

		// --- Тестируем GetTransactionsByUserID ---
		retrieved, err := repo.GetTransactionsByUserID(ctx, "user123", "")
		require.NoError(t, err)
		require.Len(t, retrieved, 1)
		assert.Equal(t, tx.UserID, retrieved[0].UserID)
		assert.Equal(t, tx.TransactionType, retrieved[0].TransactionType)
		assert.InDelta(t, tx.Amount, retrieved[0].Amount, 0.01) // для float
	})

	// --- Тестируем фильтрацию ---
	t.Run("should filter transactions by type", func(t *testing.T) {
		winTx := models.Transaction{UserID: "user456", TransactionType: "win", Amount: 200, Timestamp: time.Now()}
		betTx := models.Transaction{UserID: "user456", TransactionType: "bet", Amount: 50, Timestamp: time.Now()}

		require.NoError(t, repo.SaveTransaction(ctx, winTx))
		require.NoError(t, repo.SaveTransaction(ctx, betTx))

		// Проверяем фильтр по "win"
		wins, err := repo.GetTransactionsByUserID(ctx, "user456", "win")
		require.NoError(t, err)
		require.Len(t, wins, 1)
		assert.Equal(t, "win", wins[0].TransactionType)

		// Проверяем фильтр по "bet"
		bets, err := repo.GetTransactionsByUserID(ctx, "user456", "bet")
		require.NoError(t, err)
		require.Len(t, bets, 1)
		assert.Equal(t, "bet", bets[0].TransactionType)
	})

	// --- Тестируем GetAllTransactions ---
	t.Run("should get all transactions with filtering", func(t *testing.T) {
		// Получаем все транзакции (у нас их 3 из предыдущих тестов)
		all, err := repo.GetAllTransactions(ctx, "")
		require.NoError(t, err)
		assert.Len(t, all, 3)

		// Получаем только выигрыши (1 транзакция)
		allWins, err := repo.GetAllTransactions(ctx, "win")
		require.NoError(t, err)
		assert.Len(t, allWins, 1)
	})
}

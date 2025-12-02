package consumer

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/OlgaPie/casino-transaction-system/internal/models"
	"github.com/OlgaPie/casino-transaction-system/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	kafkatc "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestConsumerHandler_ProcessMessages(t *testing.T) {
	ctx := context.Background()

	// 1. Запуск контейнера PostgreSQL
	pgContainer, dbpool, err := setupTestDB(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %s", err)
		}
	}()

	// 2. Запуск контейнера Kafka
	kafkaContainer, err := kafkatc.Run(ctx, "confluentinc/cp-kafka:7.2.1")
	require.NoError(t, err)
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate kafka container: %s", err)
		}
	}()

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, brokers)
	kafkaBroker := brokers[0]

	// 3. Подготовка тестовых данных и зависимостей
	topic := "transactions-test"
	createTestTopic(t, kafkaBroker, topic)

	testTx := models.Transaction{
		UserID:          "test-user-consumer",
		TransactionType: models.TransactionTypeBet,
		Amount:          9990,
	}
	message, err := json.Marshal(testTx)
	require.NoError(t, err)

	// 4. Настройка нашего consumer'а
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   topic,
		GroupID: "test-group",
	})
	defer func(reader *kafka.Reader) {
		err := reader.Close()
		if err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}(reader)

	repo := repository.NewPostgresRepository(dbpool)
	handler := NewHandler(reader, repo)

	// 5. Запуск consumer'а в отдельной горутине
	handlerCtx, cancel := context.WithCancel(ctx)
	done := make(chan bool)

	go func() {
		handler.ProcessMessages(handlerCtx)
		done <- true
	}()

	// 6. Отправка тестового сообщения в Kafka
	err = writeTestMessage(t, kafkaBroker, topic, message)
	require.NoError(t, err)

	// 7. Проверка результата в базе данных
	var savedTx models.Transaction
	require.Eventually(t, func() bool {
		rows, _ := dbpool.Query(ctx, "SELECT user_id, transaction_type, amount FROM transactions WHERE user_id = $1", testTx.UserID)
		if rows.Next() {
			err = rows.Scan(&savedTx.UserID, &savedTx.TransactionType, &savedTx.Amount)
			return err == nil
		}
		return false
	}, 5*time.Second, 100*time.Millisecond, "transaction was not saved to db in time")

	require.Equal(t, testTx.UserID, savedTx.UserID)
	require.Equal(t, testTx.TransactionType, savedTx.TransactionType)
	require.Equal(t, testTx.Amount, savedTx.Amount)

	// 8. Остановка consumer'а и очистка
	cancel()
	<-done
}

func setupTestDB(ctx context.Context) (*postgres.PostgresContainer, *pgxpool.Pool, error) {
	pgContainer, err := postgres.Run(ctx, "postgres:14-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	_, err = dbpool.Exec(ctx, `CREATE TABLE transactions (
        id SERIAL PRIMARY KEY,
        transaction_id VARCHAR(255) NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('bet', 'win')),
        amount BIGINT NOT NULL,
        "timestamp" TIMESTAMPTZ NOT NULL
    );
    CREATE UNIQUE INDEX idx_transactions_transaction_id ON transactions (transaction_id);
    CREATE INDEX idx_transactions_user_id ON transactions (user_id);`)
	if err != nil {
		return nil, nil, err
	}

	return pgContainer, dbpool, nil
}

func createTestTopic(t *testing.T, broker, topic string) {
	conn, err := kafka.Dial("tcp", broker)
	require.NoError(t, err)
	defer func(conn *kafka.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Failed to close Kafka connection: %v", err)
		}
	}(conn)

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	require.NoError(t, err)
}

func writeTestMessage(_ *testing.T, broker, topic string, message []byte) error {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer func(writer *kafka.Writer) {
		err := writer.Close()
		if err != nil {
			log.Printf("Failed to close Kafka writer: %v", err)
		}
	}(writer)

	return writer.WriteMessages(context.Background(), kafka.Message{Value: message})
}

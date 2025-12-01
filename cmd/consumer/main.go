package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/OlgaPie/casino-transaction-system/internal/consumer"
	"github.com/OlgaPie/casino-transaction-system/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.Println("Starting consumer...")

	// 1. Контекст с отменой по сигналу
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	postgresDSN := os.Getenv("POSTGRES_DSN")

	// 2. БД
	dbpool, err := pgxpool.New(ctx, postgresDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Connected to PostgreSQL")

	// 3. Kafka
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaBroker},
		Topic:          "transactions",
		GroupID:        "transaction-savers",
		CommitInterval: 0,
	})
	defer func() {
		if err := kafkaReader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()
	log.Println("Connected to Kafka")

	// 4. Зависимости
	txRepo := repository.NewPostgresRepository(dbpool)
	consumerHandler := consumer.NewHandler(kafkaReader, txRepo)

	// 5. Запуск с errgroup
	g := new(errgroup.Group)

	g.Go(func() error {
		consumerHandler.ProcessMessages(ctx)
		log.Println("Consumer loop finished")
		return nil
	})

	// 6. Ожидание сигнала на завершение
	<-ctx.Done()
	stop()

	log.Println("Shutting down consumer...")

	_ = g.Wait()

	log.Println("Consumer exited properly")
}

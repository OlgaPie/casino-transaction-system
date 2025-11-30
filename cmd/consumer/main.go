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
)

func main() {
	log.Println("Starting consumer...")

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	postgresDSN := os.Getenv("POSTGRES_DSN")

	//  1. Подключение к PostgreSQL
	dbpool, err := pgxpool.New(context.Background(), postgresDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Connected to PostgreSQL")

	//  2. Настройка Kafka Reader
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaBroker},
		Topic:          "transactions",
		GroupID:        "transaction-savers", // Группа потребителей, чтобы Kafka отслеживал, что мы прочли.
		CommitInterval: 0,
	})
	defer func() {
		if err := kafkaReader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()
	log.Println("Connected to Kafka")

	//  3. Инициализация зависимостей
	// Создаем экземпляр репозитория, передавая ему пул соединений с БД.
	txRepo := repository.NewPostgresRepository(dbpool)
	// Создаем обработчик сообщений, передавая ему ридер Kafka и репозиторий.
	consumerHandler := consumer.NewHandler(kafkaReader, txRepo)

	//  4. Запуск и грациозное завершение
	ctx, cancel := context.WithCancel(context.Background())
	go consumerHandler.ProcessMessages(ctx)

	// Ждем сигнала о завершении
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down consumer...")
	cancel() // Отправляем сигнал на завершение в наш обработчик
}

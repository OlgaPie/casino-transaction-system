package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/OlgaPie/casino-transaction-system/internal/handler"
	"github.com/OlgaPie/casino-transaction-system/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Starting API server...")

	postgresDSN := os.Getenv("POSTGRES_DSN")

	//  1. Подключение к PostgreSQL
	dbpool, err := pgxpool.New(context.Background(), postgresDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Connected to PostgreSQL")

	//  2. Инициализация зависимостей
	txRepo := repository.NewPostgresRepository(dbpool)
	txHandler := handler.NewTransactionHandler(txRepo)

	//  3. Настройка роутера
	r := chi.NewRouter()
	r.Use(middleware.Logger)    // Логирование всех запросов
	r.Use(middleware.Recoverer) // Восстановление после паник

	// Определяем маршруты (endpoints).
	r.Get("/transactions", txHandler.GetAllTransactions)
	r.Get("/users/{userID}/transactions", txHandler.GetUserTransactions)

	log.Println("API server is listening on :8080")
	// Запускаем веб-сервер.
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OlgaPie/casino-transaction-system/internal/handler"
	"github.com/OlgaPie/casino-transaction-system/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Starting API server...")

	// 1. Создаем главный контекст приложения
	// Он будет отменен, когда мы получим сигнал SIGINT или SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop() // defer stop() гарантирует, что ресурсы, связанные с NotifyContext, будут освобождены.

	postgresDSN := os.Getenv("POSTGRES_DSN")

	// 2. Подключение к PostgreSQL
	dbpool, err := pgxpool.New(ctx, postgresDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Connected to PostgreSQL")

	// 3. Инициализация зависимостей
	txRepo := repository.NewPostgresRepository(dbpool)
	txHandler := handler.NewTransactionHandler(txRepo)

	// 4. Настройка роутера
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Определяем маршруты (endpoints).
	r.Get("/transactions", txHandler.GetAllTransactions)
	r.Get("/users/{userID}/transactions", txHandler.GetUserTransactions)

	// 5. Настройка и запуск сервера
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("API server is listening on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 6. Ожидание сигнала на завершение
	<-ctx.Done()

	// Восстанавливаем исходный сигнал, чтобы избежать утечек.
	stop()
	log.Println("Shutting down server...")

	// 7. Грациозное завершение сервера
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}

	log.Println("Server exited properly")
}

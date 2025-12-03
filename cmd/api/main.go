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

	// Контекст завершения приложения
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	// Health endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := dbpool.Ping(r.Context()); err != nil {
			http.Error(w, "Database not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// API endpoints
	r.Get("/transactions", txHandler.GetAllTransactions)
	r.Get("/users/{userID}/transactions", txHandler.GetUserTransactions)

	// Порт из окружения
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Запуск сервера
	go func() {
		log.Printf("API server is listening on :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}

	log.Println("Server exited properly")
}

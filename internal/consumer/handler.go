package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/OlgaPie/casino-transaction-system/internal/models"
	"github.com/OlgaPie/casino-transaction-system/internal/repository"

	"github.com/segmentio/kafka-go"
)

// MessageReader определяет минимальный интерфейс, необходимый для чтения сообщений.
type MessageReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

type Handler struct {
	reader MessageReader
	repo   repository.TransactionRepository
}

func NewHandler(reader MessageReader, repo repository.TransactionRepository) *Handler {
	return &Handler{reader: reader, repo: repo}
}

func (h *Handler) ProcessMessages(ctx context.Context) {
	for {
		msg, err := h.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("Context cancelled, stopping message processing.")
				return
			}
			log.Printf("could not read message: %v", err)
			continue
		}

		var tx models.Transaction
		if err := json.Unmarshal(msg.Value, &tx); err != nil {
			log.Printf("could not unmarshal message: %v. Value: %s", err, string(msg.Value))
			continue // Пропускаем "битое" сообщение.
		}

		// Валидация данных.
		if tx.TransactionType != models.TransactionTypeBet && tx.TransactionType != models.TransactionTypeWin {
			log.Printf("invalid transaction_type: %s for user_id: %s", tx.TransactionType, tx.UserID)
			continue // Пропускаем невалидное сообщение
		}
		if tx.Amount <= 0 {
			log.Printf("invalid amount: %.2f for user_id: %s", tx.Amount, tx.UserID)
			continue
		}

		// Устанавливаем время транзакции, если его нет в сообщении
		if tx.Timestamp.IsZero() {
			tx.Timestamp = time.Now()
		}

		if err := h.repo.SaveTransaction(ctx, tx); err != nil {
			log.Printf("could not save transaction for user %s: %v", tx.UserID, err)
			// В реальном проекте здесь может быть логика повторных попыток или
			// отправка сообщения в "очередь мертвых писем" (Dead Letter Queue).
			continue
		}

		log.Printf("Successfully processed transaction for user_id: %s, amount: %.2f", tx.UserID, tx.Amount)
	}
}

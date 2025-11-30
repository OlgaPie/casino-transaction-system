package consumer

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/OlgaPie/casino-transaction-system/internal/models"
	"github.com/OlgaPie/casino-transaction-system/internal/repository"

	"github.com/segmentio/kafka-go"
)

// MessageReader определяет минимальный интерфейс, необходимый для чтения сообщений.
type MessageReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
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
		// FetchMessage получает сообщение без автоматического коммита offset
		msg, err := h.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || err == io.EOF {
				log.Println("Context cancelled or reader closed, stopping message processing.")
				return
			}
			log.Printf("could not fetch message: %v", err)
			continue
		}

		var tx models.Transaction
		if err := json.Unmarshal(msg.Value, &tx); err != nil {
			log.Printf("could not unmarshal message: %v. Value: %s", err, string(msg.Value))
			if err := h.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("failed to commit invalid message: %v", err)
			}
			continue
		}

		// Валидация данных.
		if tx.TransactionType != models.TransactionTypeBet && tx.TransactionType != models.TransactionTypeWin {
			log.Printf("invalid transaction_type: %s for user_id: %s", tx.TransactionType, tx.UserID)
			if err := h.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("failed to commit invalid message: %v", err)
			}
			continue
		}
		if tx.Amount <= 0 {
			log.Printf("invalid amount: %d for user_id: %s", tx.Amount, tx.UserID)
			if err := h.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("failed to commit invalid message: %v", err)
			}
			continue
		}

		if tx.TransactionID == "" {
			tx.TransactionID = generateTransactionID(tx)
			log.Printf("generated transaction_id: %s for user_id: %s", tx.TransactionID, tx.UserID)
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

		// Коммитим сообщение только после успешного сохранения в БД
		if err := h.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit message after successful save: %v", err)
		}

		log.Printf("Successfully processed and committed transaction for user_id: %s, amount: %d", tx.UserID, tx.Amount)
	}
}

func generateTransactionID(tx models.Transaction) string {
	data := fmt.Sprintf("%s:%s:%d:%d", tx.UserID, tx.TransactionType, tx.Amount, tx.Timestamp.UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

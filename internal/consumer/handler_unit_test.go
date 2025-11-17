package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/OlgaPie/casino-transaction-system/internal/models"
	"github.com/OlgaPie/casino-transaction-system/internal/repository/mocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

type MockMessageReader struct {
	mock.Mock
}

func (m *MockMessageReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	if msg, ok := args.Get(0).(kafka.Message); ok {
		return msg, args.Error(1)
	}
	return kafka.Message{}, args.Error(1)
}

func TestConsumerHandler_ErrorScenarios(t *testing.T) {
	//  Тест 1: Ошибка парсинга JSON
	t.Run("should skip message on unmarshal error", func(t *testing.T) {
		mockReader := new(MockMessageReader)
		mockRepo := new(mocks.TransactionRepository)
		ctx, cancel := context.WithCancel(context.Background())

		badMessage := kafka.Message{Value: []byte("not a json")}
		mockReader.On("ReadMessage", mock.Anything).Return(badMessage, nil).Once()
		mockReader.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe() // Maybe() означает, что вызов может быть, а может и не быть

		handler := NewHandler(mockReader, mockRepo)

		go handler.ProcessMessages(ctx)

		time.Sleep(50 * time.Millisecond)
		cancel()

		mockRepo.AssertNotCalled(t, "SaveTransaction", mock.Anything, mock.Anything)
		mockReader.AssertExpectations(t)
	})

	//  Тест 2: Невалидный тип транзакции
	t.Run("should skip message with invalid transaction type", func(t *testing.T) {
		mockReader := new(MockMessageReader)
		mockRepo := new(mocks.TransactionRepository)
		ctx, cancel := context.WithCancel(context.Background())

		tx := models.Transaction{UserID: "u1", TransactionType: "refund", Amount: 100}
		msgBytes, _ := json.Marshal(tx)
		message := kafka.Message{Value: msgBytes}
		mockReader.On("ReadMessage", mock.Anything).Return(message, nil).Once()
		mockReader.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

		handler := NewHandler(mockReader, mockRepo)

		go handler.ProcessMessages(ctx)
		time.Sleep(50 * time.Millisecond)
		cancel()

		mockRepo.AssertNotCalled(t, "SaveTransaction", mock.Anything, mock.Anything)
		mockReader.AssertExpectations(t)
	})

	//  Тест 3: Невалидная сумма
	t.Run("should skip message with invalid amount", func(t *testing.T) {
		mockReader := new(MockMessageReader)
		mockRepo := new(mocks.TransactionRepository)
		ctx, cancel := context.WithCancel(context.Background())

		tx := models.Transaction{UserID: "u1", TransactionType: "bet", Amount: -50}
		msgBytes, _ := json.Marshal(tx)
		message := kafka.Message{Value: msgBytes}
		mockReader.On("ReadMessage", mock.Anything).Return(message, nil).Once()
		mockReader.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

		handler := NewHandler(mockReader, mockRepo)

		go handler.ProcessMessages(ctx)
		time.Sleep(50 * time.Millisecond)
		cancel()

		mockRepo.AssertNotCalled(t, "SaveTransaction", mock.Anything, mock.Anything)
		mockReader.AssertExpectations(t)
	})

	//  Тест 4: Ошибка сохранения в БД
	t.Run("should log error when repository fails", func(t *testing.T) {
		mockReader := new(MockMessageReader)
		mockRepo := new(mocks.TransactionRepository)
		ctx, cancel := context.WithCancel(context.Background())

		tx := models.Transaction{UserID: "u1", TransactionType: "win", Amount: 200}
		msgBytes, _ := json.Marshal(tx)
		message := kafka.Message{Value: msgBytes}
		mockReader.On("ReadMessage", mock.Anything).Return(message, nil).Once()
		mockReader.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

		mockRepo.On("SaveTransaction", mock.Anything, mock.AnythingOfType("models.Transaction")).Return(errors.New("db error")).Once()

		handler := NewHandler(mockReader, mockRepo)

		go handler.ProcessMessages(ctx)
		time.Sleep(50 * time.Millisecond)
		cancel()

		time.Sleep(50 * time.Millisecond)

		mockRepo.AssertExpectations(t)
		mockReader.AssertExpectations(t)
	})
}

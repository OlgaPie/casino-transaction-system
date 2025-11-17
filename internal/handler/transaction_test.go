package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OlgaPie/casino-transaction-system/internal/models"
	"github.com/OlgaPie/casino-transaction-system/internal/repository/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionHandler_GetUserTransactions(t *testing.T) {
	mockRepo := new(mocks.TransactionRepository)
	handler := NewTransactionHandler(mockRepo)
	router := chi.NewRouter()
	router.Get("/users/{userID}/transactions", handler.GetUserTransactions)

	t.Run("successful retrieval of user transactions", func(t *testing.T) {
		expectedTxs := []models.Transaction{
			{ID: 1, UserID: "user123", TransactionType: "bet", Amount: 10, Timestamp: time.Now()},
		}
		mockRepo.On("GetTransactionsByUserID", mock.Anything, "user123", "").
			Return(expectedTxs, nil).
			Once()

		req := httptest.NewRequest("GET", "/users/user123/transactions", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var returnedTxs []models.Transaction
		err := json.Unmarshal(rr.Body.Bytes(), &returnedTxs)
		require.NoError(t, err)

		require.Len(t, returnedTxs, 1)
		assert.Equal(t, expectedTxs[0].ID, returnedTxs[0].ID)
		assert.Equal(t, expectedTxs[0].UserID, returnedTxs[0].UserID)
		assert.Equal(t, expectedTxs[0].TransactionType, returnedTxs[0].TransactionType)
		assert.Equal(t, expectedTxs[0].Amount, returnedTxs[0].Amount)
		assert.True(t, expectedTxs[0].Timestamp.Truncate(time.Millisecond).Equal(returnedTxs[0].Timestamp.Truncate(time.Millisecond)))

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository returns an error for user transactions", func(t *testing.T) {
		mockRepo.On("GetTransactionsByUserID", mock.Anything, "user-error", "").
			Return([]models.Transaction{}, errors.New("database is down")).
			Once()

		req := httptest.NewRequest("GET", "/users/user-error/transactions", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return bad request for empty user id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users//transactions", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestTransactionHandler_GetAllTransactions(t *testing.T) {
	mockRepo := new(mocks.TransactionRepository)
	handler := NewTransactionHandler(mockRepo)
	router := chi.NewRouter()
	router.Get("/transactions", handler.GetAllTransactions)

	t.Run("successful retrieval of all transactions", func(t *testing.T) {
		expectedTxs := []models.Transaction{
			{ID: 1, UserID: "user1", TransactionType: "bet", Amount: 10, Timestamp: time.Now()},
			{ID: 2, UserID: "user2", TransactionType: "win", Amount: 50, Timestamp: time.Now()},
		}
		mockRepo.On("GetAllTransactions", mock.Anything, "").
			Return(expectedTxs, nil).
			Once()

		req := httptest.NewRequest("GET", "/transactions", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var returnedTxs []models.Transaction
		err := json.Unmarshal(rr.Body.Bytes(), &returnedTxs)
		require.NoError(t, err)

		require.Len(t, returnedTxs, len(expectedTxs))
		for i := range expectedTxs {
			assert.Equal(t, expectedTxs[i].ID, returnedTxs[i].ID)
			assert.Equal(t, expectedTxs[i].UserID, returnedTxs[i].UserID)
			assert.True(t, expectedTxs[i].Timestamp.Truncate(time.Millisecond).Equal(returnedTxs[i].Timestamp.Truncate(time.Millisecond)))
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful retrieval of filtered transactions", func(t *testing.T) {
		expectedTxs := []models.Transaction{
			{ID: 2, UserID: "user2", TransactionType: "win", Amount: 50, Timestamp: time.Now()},
		}
		mockRepo.On("GetAllTransactions", mock.Anything, "win").
			Return(expectedTxs, nil).
			Once()

		req := httptest.NewRequest("GET", "/transactions?type=win", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var returnedTxs []models.Transaction
		err := json.Unmarshal(rr.Body.Bytes(), &returnedTxs)
		require.NoError(t, err)
		require.Len(t, returnedTxs, 1)
		assert.Equal(t, expectedTxs[0].UserID, returnedTxs[0].UserID)
		assert.True(t, expectedTxs[0].Timestamp.Truncate(time.Millisecond).Equal(returnedTxs[0].Timestamp.Truncate(time.Millisecond)))

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository returns an error for all transactions", func(t *testing.T) {
		mockRepo.On("GetAllTransactions", mock.Anything, "").
			Return([]models.Transaction{}, errors.New("something went wrong")).
			Once()
		req := httptest.NewRequest("GET", "/transactions", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockRepo.AssertExpectations(t)
	})
}

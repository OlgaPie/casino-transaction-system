package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/OlgaPie/casino-transaction-system/internal/repository"

	"github.com/go-chi/chi/v5"
)

type TransactionHandler struct {
	repo repository.TransactionRepository
}

func NewTransactionHandler(repo repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

func (h *TransactionHandler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	txType := r.URL.Query().Get("type")

	transactions, err := h.repo.GetTransactionsByUserID(r.Context(), userID, txType)
	if err != nil {
		log.Printf("Error fetching transactions for user %s: %v", userID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		log.Printf("Error encoding response: %v", err)
		// Если произошла ошибка здесь, скорее всего уже поздно отправлять HTTP ошибку.
	}
}

func (h *TransactionHandler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	txType := r.URL.Query().Get("type")

	transactions, err := h.repo.GetAllTransactions(r.Context(), txType)
	if err != nil {
		log.Printf("Error fetching all transactions: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

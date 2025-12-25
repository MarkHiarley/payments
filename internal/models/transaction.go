package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID            uuid.UUID         `json:"id"`
	ExternalID    string            `json:"external_id" binding:"required"`
	FromAccountID uuid.UUID         `json:"from_account_id"`
	ToAccountID   uuid.UUID         `json:"to_account_id" binding:"required"`
	Type          TransactionType   `json:"type"`
	Amount        int64             `json:"amount" binding:"required,gt=0"`
	Currency      string            `json:"currency"`
	Status        TransactionStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

type TransactionType string

const (
	Transfer TransactionType = "TRANSFER"
	Pix      TransactionType = "PIX"
	Refund   TransactionType = "REFUND"
)

type TransactionStatus string

const (
	Pending   TransactionStatus = "PENDING"
	Completed TransactionStatus = "COMPLETED"
	Failed    TransactionStatus = "FAILED"
)

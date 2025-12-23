package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID            uuid.UUID         `json:"id"`
	ExternalID    string            `json:"external_id"`
	FromAccountID uuid.UUID         `json:"from_account_id"`
	ToAccountID   uuid.UUID         `json:"to_account_id"`
	Type          TransactionType   `json:"type"`
	Amount        int64             `json:"amount"`
	Currency      string            `json:"currency"`
	Status        TransactionStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

type TransactionType string

const (
	Transfer TransactionType = "transfer"
	Pix      TransactionType = "pix"
	Refund   TransactionType = "refund"
)

type TransactionStatus string

const (
	Pending   TransactionStatus = "pending"
	Completed TransactionStatus = "completed"
	Failed    TransactionStatus = "failed"
)

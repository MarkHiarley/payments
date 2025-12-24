package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID          uuid.UUID `json:"id"`
	Balance     int64     `json:"balance"`
	UserName    string    `json:"user_name" binding:"required"`
	UserCpfCnpj string    `json:"user_cpf_cnpj" binding:"required"`
	UserEmail   string    `json:"user_email" binding:"required"`
	Password    string    `json:"password" binding:"required,min=6"`
	Blocked     bool      `json:"blocked"`
	CreatedAt   time.Time `json:"created_at"`
}

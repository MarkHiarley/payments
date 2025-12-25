package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/usecases"
)

type TransactionController struct {
	usecase *usecases.TransactionUseCase
}

func NewTransactionController(usecase *usecases.TransactionUseCase) *TransactionController {
	return &TransactionController{usecase: usecase}
}

func (tc *TransactionController) Transfer(c *gin.Context) {

	userEmail, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated user"})
		return
	}

	var input models.Transaction
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json: " + err.Error()})
		return
	}

	if err := tc.usecase.Transfer(c.Request.Context(), &input, userEmail.(string)); err != nil {
		if errors.Is(err, usecases.ErrDuplicateTransaction) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, usecases.ErrTransactionConcluded) || errors.Is(err, usecases.ErrTransactionConcludedDB) {
			c.JSON(http.StatusOK, gin.H{"message": "transaction already processed"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "transaction processed successfully"})
}

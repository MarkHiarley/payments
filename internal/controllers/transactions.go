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

func NewTransactionController(uc *usecases.TransactionUseCase) *TransactionController {
	return &TransactionController{
		usecase: uc,
	}
}

func (tc *TransactionController) Transfer(c *gin.Context) {
	var input models.Transaction

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	if err := tc.usecase.Transfer(c.Request.Context(), &input); err != nil {

		if errors.Is(err, usecases.ErrDuplicateTransaction) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Duplicate transaction",
				"message": err.Error(),
			})
			return
		}

		if errors.Is(err, usecases.ErrTransactionConcluded) || errors.Is(err, usecases.ErrTransactionConcludedDB) {
			c.JSON(http.StatusOK, gin.H{
				"error":   "Transaction already completed",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaction processed successfully",
	})
}

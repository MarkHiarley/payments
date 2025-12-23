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

	// 1. ShouldBindJSON: Tenta converter o corpo da requisição para a struct Transaction.
	// Se o JSON estiver errado (ex: faltar campo obrigatório), ele já retorna erro.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido: " + err.Error()})
		return
	}

	// 2. Chama o UseCase passando o context do Gin
	// O Gin Context carrega informações de timeout e cancelamento
	if err := tc.usecase.Transfer(c.Request.Context(), &input); err != nil {
		// Se for transação duplicada, retorna 409 Conflict
		if errors.Is(err, usecases.ErrDuplicateTransaction) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Transação duplicada",
				"message": "Esta transação já foi processada ou está em processamento",
			})
			return
		}

		// Outros erros retornam 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Resposta de Sucesso (201 Created)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Transação processada com sucesso",
		"id":      input.ID, // Se você gerou o ID no usecase
	})
}

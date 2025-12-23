package main

import (
	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/controllers"
	db "github.com/markHiarley/payments/internal/database/postgres"
	"github.com/markHiarley/payments/internal/database/redis"
	"github.com/markHiarley/payments/internal/middleware"
	"github.com/markHiarley/payments/internal/repository"
	"github.com/markHiarley/payments/internal/usecases"
)

func main() {

	db, err := db.ConnectDB()
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	redisClient, err := redis.NewRedisClient()
	if err != nil {
		panic("failed to connect redis: " + err.Error())
	}

	r := gin.Default()
	repo := repository.NewPostgresTransactionRepository(db)
	rStore := middleware.NewRedisTransactionStore(redisClient)
	usecase := usecases.NewTransactionUsecase(repo, rStore)

	controller := controllers.NewTransactionController(usecase)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/transactions", controller.Transfer)
	}

	r.Run(":8080")

}

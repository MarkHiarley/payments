package main

import (
	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/cache"
	"github.com/markHiarley/payments/internal/controllers"
	db "github.com/markHiarley/payments/internal/database/postgres"
	"github.com/markHiarley/payments/internal/database/redis"
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
	repoTrans := repository.NewPostgresTransactionRepository(db)
	repoAcc := repository.NewPostgresAccountRepository(db)

	rStore := cache.NewRedisTransactionStore(redisClient)

	usecaseTrans := usecases.NewTransactionUsecase(repoTrans, rStore)
	usecaseAcc := usecases.NewAccountUseCase(repoAcc)

	controllerTrans := controllers.NewTransactionController(usecaseTrans)
	controllerAcc := controllers.NewAccountController(usecaseAcc)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/transactions", controllerTrans.Transfer)
		v1.POST("/accounts", controllerAcc.Create)
	}

	r.Run(":8080")

}

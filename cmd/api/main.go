package main

import (
	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/cache"
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
	repoTrans := repository.NewPostgresTransactionRepository(db)
	repoAcc := repository.NewPostgresAccountRepository(db)
	repoLog := repository.NewPostgresLoginRepository(db)

	rStore := cache.NewRedisTransactionStore(redisClient)

	usecaseTrans := usecases.NewTransactionUsecase(repoTrans, repoAcc, rStore)
	usecaseAcc := usecases.NewAccountUseCase(repoAcc)
	usecaseLog := usecases.NewLoginUseCase(repoLog)

	controllerTrans := controllers.NewTransactionController(usecaseTrans)
	controllerAcc := controllers.NewAccountController(usecaseAcc)
	controllerLog := controllers.NewLoginController(usecaseLog)

	v1 := r.Group("/api/v1")
	{
		// Rotas públicas (sem autenticação)
		v1.POST("/accounts", controllerAcc.Create)
		v1.POST("/login", controllerLog.AuthenticateUser)

		// Rotas protegidas (requerem JWT)
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			protected.POST("/transactions", controllerTrans.Transfer)
		}
	}

	r.Run(":8080")

}

package main

import (
	"github.com/Aman-Shitta/rag-redis/auth"
	"github.com/Aman-Shitta/rag-redis/controller"
	"github.com/Aman-Shitta/rag-redis/database"
	"github.com/Aman-Shitta/rag-redis/routes"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

func main() {
	// load env data
	loadEnv()
	router := gin.Default()
	// err := database.InitDB()

	fbAuthService, err := auth.NewAuthService()
	if err != nil {
		panic(err)
	}

	mongoService, err := database.NewDbService()
	if err != nil {
		panic(err)
	}

	redisService := database.NewRedRedisServie()

	userHandler := controller.NewUserController(fbAuthService, mongoService)
	chatHandler := controller.NewChatController(fbAuthService, mongoService)
	socketController := controller.NewWSController(fbAuthService, mongoService, redisService)

	if err != nil {
		panic(err)
	}

	// UserController =

	apiRoutes := router.Group("/api/v1")
	{
		routes.RegisterApiRoutes(apiRoutes, userHandler)

		routes.RegisterPrivateChatRoutes(apiRoutes, chatHandler)
		routes.RegisterWSRoutes(apiRoutes, socketController)
	}

	router.Run(":8000")
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env")
	}
}

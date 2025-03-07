package routes

import (
	"github.com/Aman-Shitta/rag-redis/controller"
	"github.com/Aman-Shitta/rag-redis/middleware"
	"github.com/gin-gonic/gin"
)

// Register routes to app
func RegisterApiRoutes(r *gin.RouterGroup, userController *controller.UserController) {
	r.POST("/login", userController.Login)
	r.POST("/register", userController.Register)
}

func RegisterPrivateChatRoutes(r *gin.RouterGroup, chatController *controller.ChatController) {
	r.Use(middleware.FirebaseJWTAuthMiddleware(chatController.AuthService))

	r.POST("/chat/create/", chatController.CreateChatRoom)
	r.GET("/chats", chatController.ListChats)
}

func RegisterWSRoutes(r *gin.RouterGroup, socketController *controller.WSController) {
	r.Use(middleware.FirebaseJWTAuthMiddleware(socketController.AuthService))

	r.GET("/chat/:chat_id", socketController.JoinChat)
	r.GET("/chat/ws", socketController.Connect)

}

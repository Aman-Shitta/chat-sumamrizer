package controller

import (
	"net/http"

	"github.com/Aman-Shitta/rag-redis/auth"
	"github.com/Aman-Shitta/rag-redis/database"
	"github.com/Aman-Shitta/rag-redis/types"
	"github.com/Aman-Shitta/rag-redis/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ChatRepository struct {
	Collection *mongo.Collection
}

type ChatController struct {
	AuthService  *auth.FbAuthService
	RedisService *database.RedisService
	ChatRepo     *ChatRepository
	UserRepo     *UserRepository
}

func NewChatController(authService *auth.FbAuthService, dbservice *database.DBservice) *ChatController {
	chatCollection := dbservice.DB.Collection("chat")
	userCollection := dbservice.DB.Collection("users")
	return &ChatController{AuthService: authService, ChatRepo: &ChatRepository{Collection: chatCollection}, UserRepo: &UserRepository{Collection: userCollection}}
}

func (cc *ChatController) ListChats(c *gin.Context) {

	query := bson.D{}
	curr, err := cc.ChatRepo.Collection.Find(c.Request.Context(), query)
	if err != nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "something went wrong", map[string]string{"error": err.Error()})
		return
	}

	var chats []types.Chat
	if err := curr.All(c.Request.Context(), &chats); err != nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "something went wrong", map[string]string{"error": err.Error()})
		return
	}
	curr.Close(c.Request.Context())

	utils.SendApiResponse(c, http.StatusOK, "chats listed successfuly", chats)

}

func (cc *ChatController) CreateChatRoom(c *gin.Context) {
	var chatData = new(types.CreateChatRequest)

	if err := c.Bind(chatData); err != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "Invalid Payload", map[string]string{"error": err.Error()})
		return
	}

	query := bson.D{{Key: "name", Value: chatData.Name}}

	result := cc.ChatRepo.Collection.FindOne(c.Request.Context(), query)

	if result.Err() == nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "chat with name already exists", nil)
		return
	} else if result.Err() != mongo.ErrNoDocuments {
		utils.SendApiResponse(c, http.StatusBadRequest, "something wrong in mongo", map[string]string{"error": result.Err().Error()})
		return
	}

	var chat = types.Chat{
		Name: chatData.Name,
	}
	_, err := cc.ChatRepo.Collection.InsertOne(c.Request.Context(), chat)
	if err != nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "something went wrong", map[string]string{"error": err.Error()})
		return
	}
	utils.SendApiResponse(c, http.StatusOK, "Chat created.", chat.Name)
}

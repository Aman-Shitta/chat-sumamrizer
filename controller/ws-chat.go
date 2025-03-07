package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	firebase_auth "firebase.google.com/go/v4/auth"
	"github.com/Aman-Shitta/rag-redis/auth"
	"github.com/Aman-Shitta/rag-redis/database"
	"github.com/Aman-Shitta/rag-redis/types"
	"github.com/Aman-Shitta/rag-redis/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type WSController struct {
	sync.Mutex
	AuthService       *auth.FbAuthService
	RedisService      *database.RedisService
	ChatRepo          *ChatRepository
	UserRepo          *UserRepository
	activeConnections map[string]*websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSController(authService *auth.FbAuthService, databaseService *database.DBservice, redisService *database.RedisService) *WSController {
	chatCollection := databaseService.DB.Collection("chat")
	userCollection := databaseService.DB.Collection("users")
	return &WSController{
		AuthService:       authService,
		RedisService:      redisService,
		ChatRepo:          &ChatRepository{Collection: chatCollection},
		UserRepo:          &UserRepository{Collection: userCollection},
		activeConnections: make(map[string]*websocket.Conn),
	}
}

func (ws *WSController) JoinChat(c *gin.Context) {

	// get chat_id for user to join in
	id, _ := c.Params.Get("chat_id")
	hexId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "invalid channel id", nil)
		return
	}

	// check if the chat to join existss in DB or not
	chatQuery := bson.D{{Key: "_id", Value: hexId}}
	res := ws.ChatRepo.Collection.FindOne(c.Request.Context(), chatQuery)
	if res.Err() != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "chat does not exist", map[string]string{"error": res.Err().Error()})
		return
	}

	// decode the chat db result to chat type
	var chat types.Chat
	if err := res.Decode(&chat); err != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "chat does not exist", map[string]string{"error": err.Error()})
		return
	}

	// get firebase token to verify if users exists or not
	idToken, _ := c.Get("FIREBASE_ID_TOKEN")

	firebaseIdToken, ok := idToken.(*firebase_auth.Token)
	if !ok {
		utils.SendApiResponse(c, http.StatusInternalServerError, "Invalid token format", nil)
		return
	}
	UID := firebaseIdToken.UID
	var user types.User
	userQuery := bson.D{{Key: "uid", Value: UID}}

	err = ws.UserRepo.Collection.FindOne(c.Request.Context(), userQuery).Decode(&user)
	if err != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "user not found", map[string]string{"error": err.Error()})
		return
	}

	// add user entry to chat DB
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "users", Value: user.ID}}}}
	ws.ChatRepo.Collection.UpdateOne(c.Request.Context(), chatQuery, update)

	// Add user to chat set in redis
	ws.RedisService.Client.SAdd(c.Request.Context(), "chat:"+chat.ID.Hex(), UID)

	// Add user chats
	ws.RedisService.Client.SAdd(c.Request.Context(), "user_chats:"+UID, chat.ID.Hex())

	if conn, exists := ws.activeConnections[UID]; exists {
		go ws.subscribeToChat(chat.ID.Hex(), conn)
	}

	utils.SendApiResponse(c, http.StatusOK, "User added", nil)
}

func (ws *WSController) Connect(c *gin.Context) {

	// check if token is present
	idToken, ok := c.Get("FIREBASE_ID_TOKEN")
	if !ok {
		utils.SendApiResponse(c, http.StatusBadRequest, "invalid stored token: token expired ?", nil)
		return
	}

	// reflect the token to token object
	firebaseToken, ok := idToken.(*firebase_auth.Token)
	if !ok {
		utils.SendApiResponse(c, http.StatusBadRequest, "invalid stored token : bad token : token expired ?", nil)
		return
	}

	// retrive firbease ID for the user
	UID := firebaseToken.UID

	// get user object from mongo based on UID
	var user types.User
	userQuery := bson.D{{Key: "uid", Value: UID}}

	singleRes := ws.UserRepo.Collection.FindOne(c.Request.Context(), userQuery)
	if singleRes.Err() != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "user not in mongo db ", map[string]string{"error": singleRes.Err().Error()})
		return
	}
	err := singleRes.Decode(&user)
	if err != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "user decoding failed", map[string]string{"error": err.Error()})
		return
	}

	// change the connection to be a websocket
	w, r := c.Writer, c.Request
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "error uoppgrading connection to socket", map[string]string{"error": err.Error()})
		return
	}

	// Add user to active connections in redis
	ws.RedisService.Client.SAdd(c.Request.Context(), "active_users", UID)

	chatIds, err := ws.RedisService.Client.SMembers(c.Request.Context(), "user_chats:"+UID).Result()
	if err != nil {
		errorMessage := map[string]string{"type": "error", "message": "Error getting chat memberships"}
		conn.WriteJSON(errorMessage)
		conn.Close() // Close connection
		return
	}
	// if no memebers in chat look from mongo db
	// to populate the chat set for user
	if len(chatIds) == 0 {
		var userChats []types.Chat
		chatQuery := bson.D{{Key: "users", Value: user.ID}}
		curr, err := ws.ChatRepo.Collection.Find(c.Request.Context(), chatQuery)
		if err != nil {
			errorMessage := map[string]string{"type": "error", "message": "Error getting chat memberships"}
			conn.WriteJSON(errorMessage)
			conn.Close()
			return
		}
		err = curr.All(c.Request.Context(), &userChats)
		if err != nil {
			errorMessage := map[string]string{"type": "error", "message": "Error getting chat memberships"}
			conn.WriteJSON(errorMessage)
			conn.Close()
			return
		}

		for _, chat := range userChats {
			chatIds = append(chatIds, chat.ID.Hex())
			ws.RedisService.Client.SAdd(c.Request.Context(), "user_chats:"+user.UID, chat.ID.Hex())
			ws.RedisService.Client.SAdd(c.Request.Context(), "chat:"+chat.ID.Hex(), user.UID)
		}
	}

	// remove any existing connections for the user
	ws.Lock()
	if old, exists := ws.activeConnections[user.UID]; exists {
		old.Close()
		// delete(ws.activeConnections, user.UID)
	}
	ws.activeConnections[user.UID] = conn
	ws.Unlock()

	// Subscribe to all chat rooms
	for _, chat := range chatIds {
		go ws.subscribeToChat(chat, conn)
	}
	go ws.handleMessages(conn, user.UID)

}

func (ws *WSController) subscribeToChat(chatID string, conn *websocket.Conn) {
	pubsub := ws.RedisService.SubscribeMessages("chat:" + chatID)
	ch := pubsub.Channel()

	// Listen for new messages
	for msg := range ch {
		fmt.Println("New message received:", msg.Payload)

		// Send message only to the subscribed WebSocket connection
		if err := conn.WriteJSON(msg.Payload); err != nil {
			fmt.Println("Error sending message:", err)
			ws.Lock()
			conn.Close()
			delete(ws.activeConnections, chatID)
			ws.Unlock()
			return
		}
	}
}

func (ws *WSController) handleMessages(conn *websocket.Conn, uid string) {
	defer func() {
		ws.Lock()
		delete(ws.activeConnections, uid)
		ws.Unlock()
		conn.Close()
		ws.RedisService.Client.SRem(context.Background(), "active_users", uid)

		// Remove user from all chats
		chatIDs, err := ws.RedisService.Client.SMembers(context.Background(), "user_chats:"+uid).Result()
		if err == nil {
			for _, chatID := range chatIDs {
				ws.RedisService.Client.SRem(context.Background(), "chat:"+chatID, uid)
			}
			// Finally, remove all chat references for this user
			ws.RedisService.Client.Del(context.Background(), "user_chats:"+uid)
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			if err == io.EOF || websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				conn.Close()
				return
			}
			conn.WriteJSON(map[string]string{"error": "something went wrong : " + err.Error()})
			return
		}

		// TODO: Process incoming message and route it accordingly
		fmt.Printf("Recieved from %s: %s\n", uid, string(msg))
		if err := ws.processMessage(msg, uid); err != nil {
			conn.WriteJSON(map[string]string{"error": "something went wrong : " + err.Error()})
		}

	}
}

func (ws *WSController) processMessage(msg []byte, senderID string) error {

	var messagePayload types.SendMessageRequest

	if err := json.Unmarshal(msg, &messagePayload); err != nil {
		fmt.Println("Unmarshaling message error  : ", err.Error())
		return fmt.Errorf("invalid message format: %s", err.Error())
	}

	// modify the senderID
	messagePayload.SenderID = senderID

	// Re-marshal the modified struct back to JSON
	updatedMsg, err := json.Marshal(messagePayload)
	if err != nil {
		fmt.Println("Marshaling modified message error:", err.Error())
		return fmt.Errorf("failed to marshal modified message: %s", err.Error())
	}

	if err := ws.RedisService.PublishMessage("chat:"+messagePayload.ChatID, updatedMsg); err != nil {
		fmt.Println("publishing error :: ", err.Error())
		return fmt.Errorf("invalid message format: %s", err.Error())
	}
	return nil
}

package types

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Message struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ChatID    bson.ObjectID `bson:"chat_id" json:"chatId"`
	SenderID  bson.ObjectID `bson:"sender_id" json:"senderId"`
	Content   string        `bson:"content" json:"content"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
}

// When a user sends a message via WebSocket
type SendMessageRequest struct {
	ChatID   string `json:"chat_id" binding:"required"`
	SenderID string `json:"sender_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// When a message is sent successfully (response to sender & broadcast to others)
type MessageResponse struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

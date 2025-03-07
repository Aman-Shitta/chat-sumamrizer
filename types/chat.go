// https://www.mongodb.com/docs/drivers/go/current/usage-examples/find/

package types

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	// "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type Chat struct {
	ID        bson.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string          `json:"name" bson:"name"`
	Users     []bson.ObjectID `json:"users,omitempty" bson:"users,omitempty"`
	CreatedAt time.Time       `json:"createdAt" bson:"createdAt"`
}

type CreateChatRequest struct {
	Name string `json:"name" binding:"required"`
}

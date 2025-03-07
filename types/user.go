package types

import "go.mongodb.org/mongo-driver/v2/bson"

type User struct {
	ID       bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UID      string        `json:"uid" bson:"uid"`
	Name     string        `json:"name" binding:"required" bson:"nam"`
	Username string        `json:"username" binding:"required" bson:"username"`
	Email    string        `json:"email" binding:"required,email" bson:"email"`
	Active   bool          `json:"is_active" bson:"is_active"`
}

type Login struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterUser struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Aman-Shitta/rag-redis/auth"
	"github.com/Aman-Shitta/rag-redis/database"
	"github.com/Aman-Shitta/rag-redis/types"
	"github.com/Aman-Shitta/rag-redis/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepository struct {
	Collection *mongo.Collection
}

type UserController struct {
	AuthService *auth.FbAuthService
	UserRepo    *UserRepository
}

func NewUserController(fbAuthService *auth.FbAuthService, dbservice *database.DBservice) *UserController {
	userCollection := dbservice.DB.Collection("users")
	return &UserController{AuthService: fbAuthService, UserRepo: &UserRepository{Collection: userCollection}}
}

func (u *UserController) Login(c *gin.Context) {

	var (
		loginPayload = new(types.Login)
		err          error
	)

	if err = c.Bind(loginPayload); err != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, fmt.Sprintf("invalid payload : %v", err), nil)
		return
	}

	query := bson.D{{Key: "email", Value: loginPayload.Email}}

	res := u.UserRepo.Collection.FindOne(context.TODO(), query)

	if res.Err() == mongo.ErrNoDocuments {
		utils.SendApiResponse(c, http.StatusBadRequest, "user not available, please register", nil)
		return
	} else if res.Err() != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, fmt.Sprintf("db error : %v", res.Err()), nil)
		return
	}

	token, err := u.AuthService.Login(*loginPayload)

	if err != nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "firebase error "+err.Error(), nil)
		return
	}

	utils.SendApiResponse(c, http.StatusOK, "Login successfull", token)

}

func (u *UserController) Register(c *gin.Context) {
	var err error
	var user = new(types.RegisterUser)

	if err := c.Bind(user); err != nil {
		utils.SendApiResponse(c, http.StatusBadRequest, "invalid payload", map[string]string{"error": err.Error()})
		return
	}

	query := bson.D{{Key: "email", Value: user.Email}}

	mgContext := context.TODO()
	var existingUser types.User

	if err := u.UserRepo.Collection.FindOne(mgContext, query).Decode(&existingUser); err == nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "User already exists", nil)
		return
	} else if err != mongo.ErrNoDocuments {
		utils.SendApiResponse(c, http.StatusBadRequest, "something wrong in mongo", map[string]string{"error": err.Error()})
		return
	}

	// register to firbase
	err = u.AuthService.Register(*user)
	if err != nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "error registering to firebase", map[string]string{"error": err.Error()})
		return
	}

	// set user active
	var ut = types.User{
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Active:   true,
	}
	_, err = u.UserRepo.Collection.InsertOne(mgContext, ut)
	if err != nil {
		utils.SendApiResponse(c, http.StatusInternalServerError, "something went wrong inserting data", map[string]string{"error": err.Error()})
		return
	}

	// sucess respones
	utils.SendApiResponse(c, http.StatusOK, "user registered successfully", map[string]string{"email": ut.Email})

}

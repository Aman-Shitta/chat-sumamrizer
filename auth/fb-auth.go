// // https://github.com/firebase/firebase-admin-go/blob/570427a0f270b9adb061f54187a2b033548c3c9e/snippets/auth.go#L82-L92
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/Aman-Shitta/rag-redis/types"
	"google.golang.org/api/option"
)

const FIREBASE_URL = "https://identitytoolkit.googleapis.com/v1/accounts"

func GetFirebaseAPIKey() string {
	return os.Getenv("FIREBASE_API_KEY")
}

// https://firebase.google.com/docs/reference/rest/auth
// Sign API to signing and register user
type FbAuthService struct {
	FBAuthApp    *firebase.App
	FBAuthClient *auth.Client
}

// Create a authservice
func NewAuthService() (*FbAuthService, error) {
	fbApp, err := initFirebasApp()

	if err != nil {
		return nil, err
	}
	client, err := fbApp.Auth(context.TODO())

	if err != nil {
		return nil, err
	}

	return &FbAuthService{
		FBAuthApp:    fbApp,
		FBAuthClient: client,
	}, nil
}

// intializes the firebase app
func initFirebasApp() (*firebase.App, error) {

	opt := option.WithCredentialsFile("<path-to-firebase-json>")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

func (s *FbAuthService) Login(loginData types.Login) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	_, err := s.FBAuthClient.GetUserByEmail(ctx, loginData.Email)

	if err != nil {
		return "", err
	}

	firbaseLoginPayload, _ := json.Marshal(map[string]string{
		"email":             loginData.Email,
		"password":          loginData.Password,
		"returnSecureToken": "true",
	})

	resp, err := http.Post(FIREBASE_URL+":signInWithPassword?key="+GetFirebaseAPIKey(), "application/json", bytes.NewBuffer(firbaseLoginPayload))
	if err != nil {
		return "", err
	}

	var fbLoginReponse *types.FirebaseLoginResponse

	if err := json.NewDecoder(resp.Body).Decode(&fbLoginReponse); err != nil {
		return "", err
	}
	resp.Body.Close()

	token := fbLoginReponse.IDToken

	return token, nil
}

func (s *FbAuthService) Register(user types.RegisterUser) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var err error

	record, err := s.FBAuthClient.GetUserByEmail(ctx, user.Email)

	if err == nil && record != nil {
		return fmt.Errorf("user with %s alread exists", user.Email)
	}

	params := (&auth.UserToCreate{}).
		Email(user.Email).
		EmailVerified(true).
		Password(user.Password)

	_, err = s.FBAuthClient.CreateUser(ctx, params)

	if err != nil {
		return fmt.Errorf("error user creating firebase :%s ", err.Error())
	}

	return nil
}

package middleware

// https://github.com/auth0/go-jwt-middleware/blob/master/examples/gin-example/middleware.go

import (
	"net/http"
	"strings"

	"github.com/Aman-Shitta/rag-redis/auth"
	"github.com/Aman-Shitta/rag-redis/utils"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader   = "AUTHORIZATION"
	wsAuthorizationHeader = "Sec-WebSocket-Protocol"
	valName               = "FIREBASE_ID_TOKEN"
)

func FirebaseJWTAuthMiddleware(authservice *auth.FbAuthService) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		authHeader := ctx.Request.Header.Get(authorizationHeader)
		var token string
		if authHeader == "" {
			token = ctx.Request.Header.Get(wsAuthorizationHeader)
		} else {
			tokenSlice := strings.Fields(authHeader)
			if len(tokenSlice) != 2 && strings.ToLower(tokenSlice[0]) != "bearer" {
				utils.SendApiResponse(ctx, http.StatusUnauthorized, "unauthorized acces: check auth token", nil)
				ctx.Abort()
				return
			}

			token = tokenSlice[1]
		}

		idToken, err := authservice.FBAuthClient.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			utils.SendApiResponse(ctx, http.StatusUnauthorized, "invalid token : "+err.Error(), nil)
			ctx.Abort()
			return
		}
		ctx.Set(valName, idToken)
		ctx.Next()

	}
}

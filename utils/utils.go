package utils

import (
	"github.com/gin-gonic/gin"
)

func SendApiResponse(c *gin.Context, status int, message string, data any) {
	response := gin.H{
		"status":  status,
		"message": message,
		"data":    data,
	}

	c.JSON(status, response)
}

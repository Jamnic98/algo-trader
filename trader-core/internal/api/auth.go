package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		query := c.Query("api_key")

		if header != "ApiKey "+expectedKey && query != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorised",
			})
			return
		}

		c.Next()
	}
}

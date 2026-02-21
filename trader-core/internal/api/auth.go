package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Authorization")

		if key != "ApiKey "+expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorised",
			})
			return
		}

		c.Next()
	}
}

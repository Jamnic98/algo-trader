package api

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth guards /api/*. Two fixes over the naive version:
//
//   - Fail closed: an empty expected key (SERVER_API_KEY unset) rejects every
//     request with 500 instead of silently allowing everything.
//   - Constant-time comparison (with a length pre-check), so the key is not
//     recoverable byte-by-byte via response timing.
//
// The key is accepted as "Authorization: ApiKey <key>" or, for legacy
// clients/SSE, as ?api_key=<key> (note: query strings can end up in logs —
// the server logger strips them).
func APIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "server misconfigured: SERVER_API_KEY is not set",
			})
			return
		}

		header := c.GetHeader("Authorization")
		query := c.Query("api_key")

		headerOK := len(header) == len("ApiKey "+expectedKey) &&
			subtle.ConstantTimeCompare([]byte(header), []byte("ApiKey "+expectedKey)) == 1
		queryOK := len(query) == len(expectedKey) &&
			subtle.ConstantTimeCompare([]byte(query), []byte(expectedKey)) == 1

		if !headerOK && !queryOK {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorised",
			})
			return
		}

		c.Next()
	}
}

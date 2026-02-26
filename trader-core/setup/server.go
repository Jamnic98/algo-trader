package setup

import (
	"net/http"
	"trader-core/internal/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitServer(cfg Config) *gin.Engine {
	engine := gin.Default()

	if cfg.Env == "local" {
		engine.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:5173"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
		}))
	}

	route := engine.Group("/api")
	route.Use(api.APIKeyAuth(cfg.ApiKey))
	{
		api.RegisterAccountRoutes(route.Group("/account"))
		api.RegisterBotRoutes(route.Group("/bots"))
		api.RegisterTradeRoutes(route.Group("/trades"))
	}

	route.GET("/health", func(c *gin.Context) {
		// Authorized → return OK
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	return engine
}

package setup

import (
	"os"
	"trader-core/internal/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitServer() *gin.Engine {
	engine := gin.Default()

	if os.Getenv("APP_ENV") == "local" {
		engine.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:5173"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
		}))
	}

	route := engine.Group("/api")
	{
		api.RegisterAccountRoutes(route.Group("/account"))
		api.RegisterBotRoutes(route.Group("/bots"))
		api.RegisterTradeRoutes(route.Group("/trades"))
	}

	return engine
}

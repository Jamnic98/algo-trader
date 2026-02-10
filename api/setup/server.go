package setup

import (
	"api/handlers"

	"github.com/gin-gonic/gin"
)

func InitServer() *gin.Engine {
  engine := gin.Default()
  api := engine.Group("/api")
  {
    handlers.RegisterBotRoutes(api.Group("/bots"))
    handlers.RegisterCandleRoutes(api.Group("/candles"))
    handlers.RegisterTradeRoutes(api.Group("/trades"))
  }

  return engine
}

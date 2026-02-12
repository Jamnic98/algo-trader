package setup

import (
	"trader-core/internal/api"

	"github.com/gin-gonic/gin"
)

func InitServer() *gin.Engine {
  engine := gin.Default()
  route := engine.Group("/api")
  {
    api.RegisterBotRoutes(route.Group("/bots"))
    api.RegisterCandleRoutes(route.Group("/candles"))
    api.RegisterTradeRoutes(route.Group("/trades"))
  }

  return engine
}

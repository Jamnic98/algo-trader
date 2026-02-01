package setup

import (
	"api/handlers"

	"github.com/gin-gonic/gin"
)

func initRouters(cfg Config) {
  router := gin.Default()
  api := router.Group("/api")
  {
    handlers.RegisterCandleRoutes(api.Group("/candles"))
    handlers.RegisterTradeRoutes(api.Group("/trades"))
  }

  router.Run(":" + cfg.Port)
}

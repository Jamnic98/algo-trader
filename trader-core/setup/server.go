package setup

import (
	"fmt"
	"time"
	"trader-core/internal/api"
	"trader-core/internal/monitoring"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitServer(cfg Config, mon *monitoring.SysMonitor) *gin.Engine {
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

	route.GET("/health/stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// send immediately on connect
		fmt.Fprintf(c.Writer, "data: {\"status\":\"ok\"}\n\n")
		c.Writer.Flush()

		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(c.Writer, "data: {\"status\":\"ok\"}\n\n")
				c.Writer.Flush()
			case <-c.Request.Context().Done():
				return
			}
		}
	})

	return engine
}

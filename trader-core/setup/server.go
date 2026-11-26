package setup

import (
	"fmt"
	"time"
	"trader-core/internal/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// requestLogger logs like gin.Default but never writes the query string:
// legacy clients may pass ?api_key=..., which must not land in logs.
func requestLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %v | %3d | %13v | %s  %s\n",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.Method,
				param.Path, // path only — query may contain the API key
			)
		},
	})
}

func InitServer(cfg Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), requestLogger())

	if cfg.Env == "local" {
		engine.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:5173"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
		}))
	}

	// plain unauthenticated health check for load balancers / docker healthchecks
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})

	route := engine.Group("/api")
	route.Use(api.APIKeyAuth(cfg.ApiKey))
	{
		api.RegisterMetricsRoutes(route.Group("/metrics"))
		api.RegisterAccountRoutes(route.Group("/account"))
		api.RegisterBotRoutes(route.Group("/bots"))
		api.RegisterDiagnosticsRoutes(route.Group("/diagnostics"))
		api.RegisterPriceRoutes(route.Group("/price"))
		api.RegisterStrategyRoutes(route.Group("/strategies"))
		api.RegisterBacktestRoutes(route.Group("/backtests"), cfg.BacktesterURL)
		api.RegisterFillRoutes(route.Group("/fills"))
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

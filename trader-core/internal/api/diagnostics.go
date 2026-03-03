package api

import (
	"net/http"

	"trader-core/internal/monitoring"

	"github.com/gin-gonic/gin"
)

// InitDiagnosticsAPI registers the /diagnostics route
func InitDiagnosticsAPI(router *gin.Engine, mon *monitoring.SysMonitor) {
	router.GET("/api/diagnostics", func(c *gin.Context) {
		c.JSON(http.StatusOK, mon.GetStats())
	})
}

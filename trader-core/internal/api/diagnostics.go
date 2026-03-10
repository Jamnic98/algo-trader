package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"trader-core/internal/monitoring"

	"github.com/gin-gonic/gin"
)

type DiagnosticsPayload struct {
	Stats   monitoring.SystemStats    `json:"stats"`
	History []monitoring.HistoryPoint `json:"history"`
}

func InitDiagnosticsAPI(router *gin.Engine, mon *monitoring.SysMonitor) {
	router.GET("/api/diagnostics", func(c *gin.Context) {
		c.JSON(http.StatusOK, mon.GetStats())
	})
	router.GET("/api/diagnostics/stream", DiagnosticsSSE(mon))
}

func DiagnosticsSSE(mon *monitoring.SysMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		sendDiagnostics(c, mon)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sendDiagnostics(c, mon)
			case <-c.Request.Context().Done():
				return
			}
		}
	}
}

func sendDiagnostics(c *gin.Context, mon *monitoring.SysMonitor) {
	payload := DiagnosticsPayload{
		Stats:   mon.GetStats(),
		History: mon.GetHistory(),
	}
	data, _ := json.Marshal(payload)
	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}

func getStaticDiagnostics(mon *monitoring.SysMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, mon.GetStats())
	}
}

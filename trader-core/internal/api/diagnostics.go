package api

import (
	"encoding/json"
	"fmt"
	"time"

	"trader-core/internal/monitoring"

	"github.com/gin-gonic/gin"
)

type DiagnosticsPayload struct {
	Stats   monitoring.SystemStats    `json:"stats"`
	History []monitoring.HistoryPoint `json:"history"`
}

var sysmon *monitoring.SysMonitor

func InitDiagnosticsAPI(mon *monitoring.SysMonitor) {
	sysmon = mon
}

func RegisterDiagnosticsRoutes(rg *gin.RouterGroup) {
	rg.GET("", getStaticDiagnostics)
	rg.GET("/stream", DiagnosticsSSE)
}

func getStaticDiagnostics(c *gin.Context) {
	c.JSON(200, sysmon.GetStats())
}

func DiagnosticsSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	sendDiagnostics(c)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sendDiagnostics(c)
		case <-c.Request.Context().Done():
			return
		}
	}
}

func sendDiagnostics(c *gin.Context) {
	payload := DiagnosticsPayload{
		Stats:   sysmon.GetStats(),
		History: sysmon.GetHistory(),
	}
	data, _ := json.Marshal(payload)
	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}

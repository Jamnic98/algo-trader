package api

import (
	"encoding/json"
	"net/http"
	"time"

	"trader-core/internal/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// InitDiagnosticsAPI registers the /diagnostics route
func InitDiagnosticsAPI(router *gin.Engine, mon *monitoring.SysMonitor) {
	router.GET("/api/diagnostics", func(c *gin.Context) {
		c.JSON(http.StatusOK, mon.GetStats())
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten this for production
	},
}

// DiagnosticsWS streams SystemStats to the client every 5s.
func DiagnosticsWS(mon *monitoring.SysMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		if err := sendStats(conn, mon); err != nil {
			return
		}

		for {
			select {
			case <-c.Request.Context().Done():
				return
			case <-ticker.C:
				if err := sendStats(conn, mon); err != nil {
					return
				}
			}
		}
	}
}

func sendStats(conn *websocket.Conn, mon *monitoring.SysMonitor) error {
	b, err := json.Marshal(mon.GetStats())
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, b)
}

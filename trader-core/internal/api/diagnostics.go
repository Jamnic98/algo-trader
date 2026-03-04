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

func DiagnosticsWS(mon *monitoring.SysMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		done := make(chan struct{})

		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			return nil
		})

		// read loop — only purpose is to detect disconnect via pong
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		if err := sendStats(conn, mon); err != nil {
			return
		}

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
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

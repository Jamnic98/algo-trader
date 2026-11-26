package api

import (
	"net/http"
	"time"

	"trader-core/internal/db"
	"trader-core/internal/db/models"

	"github.com/gin-gonic/gin"
)

func RegisterMetricsRoutes(rg *gin.RouterGroup) {
	rg.GET("", getMetricsHandler)
}

// getMetricsHandler returns a small ops snapshot: bot counts, fills today,
// and account summaries. Fills count needs the DB; without it reports -1.
func getMetricsHandler(c *gin.Context) {
	activeBotsMu.RLock()
	total, running := len(activeBots), 0
	for _, b := range activeBots {
		if b.Status == "trading" {
			running++
		}
	}
	activeBotsMu.RUnlock()

	fillsToday := -1
	if db.DB != nil {
		startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
		var n int64
		if err := db.DB.Model(&models.Fill{}).Where("timestamp >= ?", startOfDay).Count(&n).Error; err == nil {
			fillsToday = int(n)
		}
	}

	accounts := []any{}
	if runtime != nil && runtime.Accounts != nil {
		for _, a := range runtime.Accounts.List() {
			accounts = append(accounts, a)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"time":        time.Now().UTC().Format(time.RFC3339),
		"bots_total":  total,
		"bots_running": running,
		"fills_today": fillsToday,
		"accounts":    accounts,
	})
}

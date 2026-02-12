package api

import (
	"errors"
	"net/http"
	"strconv"

	"trader-core/internal/db"
	"trader-core/internal/db/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register routes for bots
func RegisterBotRoutes(rg *gin.RouterGroup) {
    rg.GET("/", getBotsHandler)
    rg.GET("/:id/", getBotByIdHandler)
    rg.POST("/", createBotHandler)
    rg.DELETE("/:id/", deleteBotHandler)
}

func getBotsHandler(c *gin.Context) {
    var bots []models.Bot

    if err := db.DB.Find(&bots).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"bots": bots})
}

func getBotByIdHandler(c *gin.Context) {
    var bot models.Bot

    idStr := c.Param("id")

    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    result := db.DB.First(&bot, uint(id))
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
        return
    }
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"bot": bot})
}

func createBotHandler(c *gin.Context) {
    bot := models.Bot{}

    if err := db.DB.Create(&bot).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": bot.ID})
}

func deleteBotHandler(c *gin.Context) {
    idStr := c.Param("id")

    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    result := db.DB.Delete(&models.Bot{}, uint(id))

    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
        return
    }

    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
        return
    }

    c.Status(http.StatusNoContent)
}

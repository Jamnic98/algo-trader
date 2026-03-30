package api

import (
	"encoding/json"
	"net/http"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/strategies"

	"github.com/gin-gonic/gin"
)

func RegisterStrategyRoutes(rg *gin.RouterGroup) {
	rg.GET("", getStrategiesHandler)
	rg.GET("/:slug", getStrategyHandler)
	rg.GET("/schemas", getSchemasHandler)
	rg.POST("", createStrategyHandler)
}

func getSchemasHandler(c *gin.Context) {
	c.JSON(http.StatusOK, strategies.Schemas())
}

func getStrategiesHandler(c *gin.Context) {
	query := db.DB.Model(&models.Strategy{})

	var strategies []models.Strategy
	if err := query.
		// Order("timestamp DESC").
		Find(&strategies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	strategiesDTO := make([]dto.StrategyDTO, len(strategies))
	for i, s := range strategies {
		strategiesDTO[i] = dto.StrategyToDTO(&s)
	}

	c.JSON(http.StatusOK, gin.H{"strategies": strategiesDTO})
}

func getStrategyHandler(c *gin.Context) {
	slug := c.Param("slug")
	var strategy models.Strategy
	if err := db.DB.First(&strategy, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}
	c.JSON(http.StatusOK, dto.StrategyToDTO(&strategy))
}

type CreateStrategyRequest struct {
	Slug        string                     `json:"slug"         binding:"required"`
	DisplayName string                     `json:"display_name" binding:"required"`
	Threshold   float64                    `json:"threshold"`
	Nodes       []strategies.CompositeNode `json:"nodes" binding:"required,min=1"`
}

func createStrategyHandler(c *gin.Context) {
	var req CreateStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := strategies.CompositeConfig{
		Threshold: req.Threshold,
		Nodes:     req.Nodes,
	}
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode config"})
		return
	}

	strategy := models.Strategy{
		Slug:          req.Slug,
		DisplayName:   req.DisplayName,
		IsComposite:   true,
		DefaultConfig: cfgJSON,
		Nodes:         []byte("[]"), // kept for backwards compat, real data is in DefaultConfig
	}

	if err := db.DB.Create(&strategy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.StrategyToDTO(&strategy))
}

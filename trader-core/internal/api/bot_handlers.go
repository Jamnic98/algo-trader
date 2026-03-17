package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"trader-core/internal/bot"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func calcBotPositions(botID string) ([]dto.PositionDTO, error) {
	var trades []models.Trade
	if err := db.DB.Where("bot_id = ?", botID).Order("timestamp ASC").Find(&trades).Error; err != nil {
		return nil, err
	}

	type state struct {
		qty         decimal.Decimal
		totalSpent  decimal.Decimal // sum of buy notionals
		totalBought decimal.Decimal // total qty bought
		totalFees   decimal.Decimal
		realised    decimal.Decimal
	}

	ps := make(map[string]*state)

	priceScale := decimal.NewFromInt(1e8)
	qtyScale := decimal.NewFromInt(1e6)
	feeScale := decimal.NewFromInt(1e8)

	for _, t := range trades {
		if _, ok := ps[t.Symbol]; !ok {
			ps[t.Symbol] = &state{}
		}
		s := ps[t.Symbol]

		price := decimal.NewFromInt(t.PriceInt).Div(priceScale)
		qty := decimal.NewFromInt(t.QuantityInt).Div(qtyScale)
		fee := decimal.NewFromInt(t.FeeInt).Div(feeScale)
		notional := price.Mul(qty)

		s.totalFees = s.totalFees.Add(fee)

		switch t.Side {
		case "BUY":
			s.qty = s.qty.Add(qty)
			s.totalSpent = s.totalSpent.Add(notional)
			s.totalBought = s.totalBought.Add(qty)

		case "SELL":
			if s.totalBought.IsPositive() {
				avgEntry := s.totalSpent.Div(s.totalBought)
				s.realised = s.realised.Add(price.Sub(avgEntry).Mul(qty))
			}
			s.qty = s.qty.Sub(qty)
			s.totalSpent = s.totalSpent.Sub(price.Mul(qty))
			if s.totalBought.IsPositive() {
				s.totalBought = s.totalBought.Sub(qty)
			}
		}
	}

	result := make([]dto.PositionDTO, 0, len(ps))
	for symbol, s := range ps {
		avgEntry := decimal.Zero
		if s.totalBought.IsPositive() {
			avgEntry = s.totalSpent.Div(s.totalBought)
		}
		result = append(result, dto.PositionDTO{
			Symbol:     symbol,
			Qty:        s.qty.StringFixed(6),
			AvgEntry:   avgEntry.StringFixed(8),
			TotalSpent: s.totalSpent.StringFixed(8),
			TotalFees:  s.totalFees.StringFixed(8),
			Realised:   s.realised.StringFixed(8),
		})
	}

	return result, nil
}

var (
	runtime      *bot.Runtime
	activeBotsMu sync.RWMutex
	activeBots   = make(map[string]*bot.Bot)
)

func InitBotAPI(rt *bot.Runtime) {
	runtime = rt

	cfgs, err := runtime.LoadBots()
	if err != nil {
		log.Printf("failed to load bots: %v", err)
		return
	}
	for _, cfg := range cfgs {
		b, err := runtime.RestoreBot(cfg)
		if err != nil {
			log.Printf("failed to restore bot %s: %v", cfg.ID, err)
			continue
		}
		activeBotsMu.Lock()
		activeBots[b.ID] = b
		activeBotsMu.Unlock()
	}
}

func RegisterBotRoutes(rg *gin.RouterGroup) {
	rg.GET("", getBotsHandler)
	rg.GET("/:id", getBotByIDHandler)
	rg.GET("/:id/trades", getBotTradesHandler)
	rg.GET("/:id/trades/stream", streamBotTradesHandler)
	rg.GET("/:id/logs/stream", streamBotLogsHandler)
	rg.GET("/:id/positions", getBotPositionsHandler)
	rg.GET("/:id/candles/stream", streamBotCandlesHandler)
	rg.GET("/:id/ticks/stream", streamBotTicksHandler)

	rg.POST("/:id/start", startBotHandler)
	rg.POST("/:id/stop", stopBotHandler)
	rg.POST("", createBotHandler)
	rg.DELETE("/:id", deleteBotHandler)
}

func getBotByIDHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func getBotsHandler(c *gin.Context) {
	includeDeleted := c.Query("deleted") == "true"

	var cfgs []bot.BotConfig
	if includeDeleted {
		runtime.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&cfgs)
		dtos := make([]BotDTO, len(cfgs))
		for i, cfg := range cfgs {
			dtos[i] = configToDTO(cfg)
		}
		c.JSON(http.StatusOK, gin.H{"bots": dtos})
	} else {
		activeBotsMu.RLock()
		bots := make([]*bot.Bot, 0, len(activeBots))
		for _, b := range activeBots {
			bots = append(bots, b)
		}
		activeBotsMu.RUnlock()

		dtos := make([]BotDTO, len(bots))
		for i, b := range bots {
			dtos[i] = botToDTO(b)
		}
		c.JSON(http.StatusOK, gin.H{"bots": dtos})
	}
}

func getBotTradesHandler(c *gin.Context) {
	botId := c.Param("id")
	if getBot(botId) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	dtos, total, totalPages, err := fetchBotTrades(botId, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dtos,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func getBotPositionsHandler(c *gin.Context) {
	botId := c.Param("id")
	if getBot(botId) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "positions not found"})
		return
	}

	positions, err := calcBotPositions(botId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"positions": positions})
}

func streamBotTradesHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	ch := b.TradeBroadcaster.Subscribe()
	defer b.TradeBroadcaster.Unsubscribe(ch)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for {
		select {
		case trade := <-ch:
			data, _ := json.Marshal(trade)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func streamBotLogsHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	snapshot, ch := b.Logger.Subscribe()
	defer b.Logger.Unsubscribe(ch)

	// send catch-up buffer first
	for _, entry := range snapshot {
		data, _ := json.Marshal(entry)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	}
	c.Writer.Flush()

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(entry)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func streamBotCandlesHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ch := b.CandleBroadcaster.Subscribe()
	defer b.CandleBroadcaster.Unsubscribe(ch)

	// send existing candles as snapshot first
	for _, candle := range b.Candles {
		data, _ := json.Marshal(models.CandleToDTO(candle))
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	}
	c.Writer.Flush()

	for {
		select {
		case candle := <-ch:
			data, _ := json.Marshal(models.CandleToDTO(candle))
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func streamBotTicksHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ch := b.TickBroadcaster.Subscribe()
	defer b.TickBroadcaster.Unsubscribe(ch)

	for {
		select {
		case candle := <-ch:
			data, _ := json.Marshal(models.CandleToDTO(candle))
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func createBotHandler(c *gin.Context) {
	var req struct {
		Base     string      `json:"base"`
		Quote    string      `json:"quote"`
		Interval string      `json:"interval"`
		Lookback string      `json:"lookback"`
		Quantity string      `json:"quantity"`
		Mode     bot.BotMode `json:"mode"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Base == "" || req.Quote == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base and quote are required"})
		return
	}

	cfg, err := parseBotCreateRequest(req.Base, req.Quote, req.Interval, req.Lookback, req.Quantity, req.Mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := runtime.CreateBot(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeBotsMu.Lock()
	activeBots[b.ID] = b
	activeBotsMu.Unlock()

	c.JSON(http.StatusCreated, gin.H{"bot": botToDTO(b)})
}

func startBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	if b.Status == bot.BotCreated {
		if err := runtime.AttachBot(b); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
	} else if err := b.Start(); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func stopBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	if err := runtime.DetachBot(b); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func deleteBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	if b.Status == bot.BotRunning {
		runtime.DetachBot(b)
	}

	if err := runtime.DeleteBot(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeBotsMu.Lock()
	delete(activeBots, b.ID)
	activeBotsMu.Unlock()

	c.Status(http.StatusNoContent)
}

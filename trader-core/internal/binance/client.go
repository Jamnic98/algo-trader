package binance

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	url         string
	conn        *websocket.Conn
	send        chan any
	recv        chan []byte
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	connected   atomic.Bool
	lastMessage time.Time
}

// NewClient constructor
func NewClient(ctx context.Context, url string) *Client {
	cctx, cancel := context.WithCancel(ctx)

	return &Client{
		url:    url,
		send:   make(chan any, 10),
		recv:   make(chan []byte, 10),
		ctx:    cctx,
		cancel: cancel,
	}
}

func (c *Client) Run() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		c.connected.Store(false)
		return err
	}

	// Add pong handler
	conn.SetPongHandler(func(appData string) error {
		c.mu.Lock()
		c.lastMessage = time.Now()
		c.mu.Unlock()
		return nil
	})

	c.mu.Lock()
	c.conn = conn
	c.lastMessage = time.Now()
	c.mu.Unlock()

	c.connected.Store(true)
	go c.readLoop()
	go c.writeLoop()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				c.mu.Lock()
				err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
				c.mu.Unlock()

				if err != nil {
					log.Println("ping error:", err)
					c.connected.Store(false)
					return
				}
			}
		}
	}()

	return nil
}

// readLoop updates lastMessage on each message
func (c *Client) readLoop() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}

		// update lastMessage timestamp
		c.mu.Lock()
		c.lastMessage = time.Now()
		c.mu.Unlock()

		select {
		case c.recv <- msg:
		default:
			log.Println("recv channel full, dropping message")
		}
	}
}

// writeLoop unchanged
func (c *Client) writeLoop() {
	for {
		select {
		case msg := <-c.send:
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Println("write error:", err)
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// Messages channel
func (c *Client) Messages() <-chan []byte {
	return c.recv
}

// Subscribe/Unsubscribe unchanged
func (c *Client) Subscribe(streams ...string) {
	c.send <- map[string]any{
		"method": "SUBSCRIBE",
		"params": streams,
		"id":     time.Now().Unix(),
	}
}

func (c *Client) Unsubscribe(streams ...string) {
	c.send <- map[string]any{
		"method": "UNSUBSCRIBE",
		"params": streams,
		"id":     time.Now().Unix(),
	}
}

func (c *Client) IsAlive() bool {
	if !c.connected.Load() {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lastMessage.IsZero() {
		return true
	}

	// 15s timeout for ping/pong
	if time.Since(c.lastMessage) > 15*time.Second {
		return false
	}

	return true
}

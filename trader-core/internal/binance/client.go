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
	c.mu.Lock()
	c.conn = conn
	c.lastMessage = time.Now()
	c.mu.Unlock()

	c.connected.Store(true)
	go c.readLoop()
	go c.writeLoop()

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

	// If we’ve received at least one message, check latency
	if !c.lastMessage.IsZero() && time.Since(c.lastMessage) > 30*time.Second {
		return false
	}

	return true
}

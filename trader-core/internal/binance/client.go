package binance

import (
	"context"
	"fmt"
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
	subs        map[string]bool // active stream subscriptions, replayed on (re)connect
	connected   atomic.Bool
	started     atomic.Bool
	lastMessage time.Time
}

// var (not const) so tests/adapters can point it at a local server
var binanceWSURL = "wss://stream.binance.com:443/ws"

// NewClient constructor
func NewClient(ctx context.Context) *Client {
	cctx, cancel := context.WithCancel(ctx)

	return &Client{
		url:    binanceWSURL,
		send:   make(chan any, 10),
		recv:   make(chan []byte, 10),
		ctx:    cctx,
		cancel: cancel,
		subs:   make(map[string]bool),
	}
}

func (c *Client) StartWithReconnect(onFatal func(string)) error {
	log.Println("connecting to Binance websocket...")
	if err := c.run(); err != nil {
		return err
	}

	go func() {
		const maxReconnects = 3
		const reconnectDelay = 2 * time.Second
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		reconnectAttempts := 0

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				if !c.Started() {
					continue
				}
				if c.IsAlive() {
					reconnectAttempts = 0
					continue
				}
				reconnectAttempts++
				if reconnectAttempts > maxReconnects {
					if onFatal != nil {
						onFatal("CRITICAL: Binance WS could not reconnect after max attempts")
					}
					// do NOT cancel the client ctx here: StartWithRetry relies on
					// being able to re-dial after a fatal disconnect.
					return
				}
				log.Printf("WARNING: Binance WS disconnected, attempt %d/%d", reconnectAttempts, maxReconnects)
				if err := c.run(); err != nil {
					log.Printf("WARNING: reconnect failed: %v", err)
				} else {
					log.Println("Binance websocket reconnected")
					reconnectAttempts = 0
				}
				time.Sleep(reconnectDelay)
			}
		}
	}()

	return nil
}

func (c *Client) run() error {
	// dial the current package URL so tests/adapters can repoint between retries
	conn, _, err := websocket.DefaultDialer.Dial(binanceWSURL, nil)
	if err != nil {
		c.connected.Store(false)
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.lastMessage = time.Now()
	// replay subscriptions on every (re)connect — a fresh connection has none
	subs := make([]string, 0, len(c.subs))
	for s := range c.subs {
		subs = append(subs, s)
	}
	c.mu.Unlock()

	if len(subs) > 0 {
		c.send <- map[string]any{
			"method": "SUBSCRIBE",
			"params": subs,
			"id":     time.Now().Unix(),
		}
		log.Printf("resubscribed to %d streams after connect", len(subs))
	}

	c.connected.Store(true)
	c.started.Store(true)

	// Pong updates lastMessage
	conn.SetPongHandler(func(appData string) error {
		c.mu.Lock()
		c.lastMessage = time.Now()
		c.mu.Unlock()
		return nil
	})

	go c.readLoop()
	go c.writeLoop()
	go c.pingLoop()

	return nil
}

// in binance/client.go
func (c *Client) Started() bool {
	return c.started.Load()
}

func (c *Client) pingLoop() {
	ticker := time.NewTicker(3 * time.Minute)
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
}

func (c *Client) IsAlive() bool {
	if !c.connected.Load() {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastMessage.IsZero() {
		return false
	}
	return time.Since(c.lastMessage) <= 3*time.Minute
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

// StartWithRetry connects and keeps reconnecting until ctx is cancelled.
// Neither an initial dial failure nor a later fatal disconnect is fatal for
// the caller: notify is invoked and the connect is retried after retryDelay.
// Use this in long-running servers where market data must not take the API down.
func (c *Client) StartWithRetry(ctx context.Context, retryDelay time.Duration, notify func(string)) {
	for {
		fatal := make(chan string, 1)
		err := c.StartWithReconnect(func(msg string) {
			if notify != nil {
				notify(msg)
			}
			select {
			case fatal <- msg:
			default:
			}
		})
		if err != nil {
			if notify != nil {
				notify(fmt.Sprintf("binance ws connect failed: %v", err))
			}
			if !sleepOrDone(ctx, retryDelay) {
				return
			}
			continue
		}
		// connected; wait for shutdown or a fatal disconnect, then redial
		select {
		case <-ctx.Done():
			return
		case <-fatal:
			if !sleepOrDone(ctx, retryDelay) {
				return
			}
		}
	}
}

func sleepOrDone(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// Subscribe records the streams (replayed on reconnect) and sends the
// SUBSCRIBE if the socket is up.
func (c *Client) Subscribe(streams ...string) {
	c.mu.Lock()
	for _, s := range streams {
		c.subs[s] = true
	}
	c.mu.Unlock()

	if !c.Started() {
		log.Println("WS not started yet, will subscribe on connect")
		return
	}

	c.send <- map[string]any{
		"method": "SUBSCRIBE",
		"params": streams,
		"id":     time.Now().Unix(),
	}
}

func (c *Client) Unsubscribe(streams ...string) {
	c.mu.Lock()
	for _, s := range streams {
		delete(c.subs, s)
	}
	c.mu.Unlock()

	if !c.Started() {
		return
	}
	c.send <- map[string]any{
		"method": "UNSUBSCRIBE",
		"params": streams,
		"id":     time.Now().Unix(),
	}
}

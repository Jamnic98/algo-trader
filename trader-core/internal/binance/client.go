package binance

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
  url   string
  conn  *websocket.Conn

  send chan any
  recv chan []byte

  ctx    context.Context
  cancel context.CancelFunc

  mu sync.Mutex
}


func NewClient(ctx context.Context, url string) *Client {
  cctx, cancel := context.WithCancel(ctx)

  return &Client{
    url:    url,
    send:   make(chan any, 100),
    recv:   make(chan []byte, 1000),
    ctx:    cctx,
    cancel: cancel,
  }
}


func (c *Client) Run() error {
  conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
  if err != nil {
    return err
  }

  c.mu.Lock()
  c.conn = conn
  c.mu.Unlock()

  go c.readLoop()
  go c.writeLoop()

  log.Println("Connected to Binance WS")
  return nil
}


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


func (c *Client) readLoop() {
  for {
    _, msg, err := c.conn.ReadMessage()
    if err != nil {
      log.Println("read error:", err)
      return
    }

    select {
    case c.recv <- msg:
    default:
      log.Println("recv channel full, dropping message")
    }
  }
}


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


func (c *Client) Messages() <-chan []byte {
  return c.recv
}

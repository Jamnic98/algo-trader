package binance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestStartWithRetryConnectsAfterFailures: the client must keep retrying when
// the initial dial fails (e.g. exchange unreachable at boot) and connect once
// the endpoint appears — instead of giving up.
func TestStartWithRetryConnectsAfterFailures(t *testing.T) {
	orig := binanceWSURL
	t.Cleanup(func() { binanceWSURL = orig })

	// nothing listening on this port -> dial fails
	binanceWSURL = "ws://127.0.0.1:1"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := NewClient(ctx)

	var mu sync.Mutex
	notes := 0
	done := make(chan struct{})
	go func() {
		c.StartWithRetry(ctx, 20*time.Millisecond, func(string) {
			mu.Lock()
			notes++
			mu.Unlock()
		})
		close(done)
	}()

	// let it fail a few times
	time.Sleep(120 * time.Millisecond)
	mu.Lock()
	if notes == 0 {
		t.Fatal("expected notify calls for failed dials")
	}
	mu.Unlock()

	// now bring a real WS endpoint up
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()
	binanceWSURL = "ws" + strings.TrimPrefix(srv.URL, "http")

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && !c.Started() {
		time.Sleep(10 * time.Millisecond)
	}
	if !c.Started() {
		t.Fatal("client never connected after retries")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("StartWithRetry did not return after ctx cancel")
	}
}


// TestSubscribeReplayedOnConnect: streams subscribed before/while the socket
// is down must be replayed when (re)connecting, otherwise bots silently stop
// receiving candles after a WS reconnect.
func TestSubscribeReplayedOnConnect(t *testing.T) {
	orig := binanceWSURL
	t.Cleanup(func() { binanceWSURL = orig })

	subscribed := make(chan string, 10)
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			var msg map[string]any
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			if msg["method"] == "SUBSCRIBE" {
				for _, p := range msg["params"].([]any) {
					subscribed <- p.(string)
				}
			}
		}
	}))
	defer srv.Close()
	binanceWSURL = "ws" + strings.TrimPrefix(srv.URL, "http")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := NewClient(ctx)

	// subscribe BEFORE connecting — must be replayed by run()
	c.Subscribe("btcusdt@kline_1m")

	if err := c.run(); err != nil {
		t.Fatalf("run: %v", err)
	}

	select {
	case s := <-subscribed:
		if s != "btcusdt@kline_1m" {
			t.Fatalf("replayed sub = %q", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("subscription was not replayed on connect")
	}
}

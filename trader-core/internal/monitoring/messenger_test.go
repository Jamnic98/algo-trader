package monitoring

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type telegramCall struct {
	chatID string
	text   string
}

func fakeTelegram(t *testing.T, status int, body string) (*httptest.Server, func() []telegramCall) {
	t.Helper()
	var mu sync.Mutex
	var calls []telegramCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		mu.Lock()
		calls = append(calls, telegramCall{chatID: r.Form.Get("chat_id"), text: r.Form.Get("text")})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	old := telegramAPIBase
	telegramAPIBase = srv.URL
	t.Cleanup(func() { telegramAPIBase = old; srv.Close() })
	return srv, func() []telegramCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]telegramCall{}, calls...)
	}
}

func newTestMessenger() *Messenger {
	return &Messenger{
		BotToken:      "token",
		ChatID:        "main-chat",
		ErrorChatID:   "error-chat",
		Messages:      make(chan string, 10),
		ErrorMessages: make(chan string, 10),
		Quit:          make(chan struct{}),
	}
}

func waitForCalls(t *testing.T, getCalls func() []telegramCall, n int) []telegramCall {
	t.Helper()
	for i := 0; i < 50; i++ {
		if calls := getCalls(); len(calls) >= n {
			return calls
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d telegram calls, got %v", n, getCalls())
	return nil
}

func TestMessengerRoutesErrorChat(t *testing.T) {
	_, getCalls := fakeTelegram(t, http.StatusOK, `{"ok":true}`)
	m := newTestMessenger()
	go m.Run()
	defer close(m.Quit)

	m.Notify("hello")
	m.NotifyError("boom")

	calls := waitForCalls(t, getCalls, 2)
	byText := map[string]string{}
	for _, c := range calls {
		byText[c.text] = c.chatID
	}
	if byText["hello"] != "main-chat" {
		t.Fatalf("general message went to chat %q, want main-chat", byText["hello"])
	}
	if byText["boom"] != "error-chat" {
		t.Fatalf("error message went to chat %q, want error-chat", byText["boom"])
	}
}

func TestMessengerErrorFallsBackToMainChat(t *testing.T) {
	_, getCalls := fakeTelegram(t, http.StatusOK, `{"ok":true}`)
	m := newTestMessenger()
	m.ErrorChatID = "" // no dedicated error chat configured
	go m.Run()
	defer close(m.Quit)

	m.NotifyError("boom")
	calls := waitForCalls(t, getCalls, 1)
	if calls[0].chatID != "main-chat" {
		t.Fatalf("error fallback went to chat %q, want main-chat", calls[0].chatID)
	}
}

func TestMessengerSendFailureIsLoggedNotSilent(t *testing.T) {
	fakeTelegram(t, http.StatusBadRequest, `{"ok":false,"description":"chat not found"}`)

	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	m := newTestMessenger()
	m.send(&http.Client{Timeout: time.Second}, m.ChatID, "hello")

	if !bytes.Contains(buf.Bytes(), []byte("chat not found")) {
		t.Fatalf("expected failure to be logged, got: %s", buf.String())
	}
}

func TestMessengerNotifyAfterQuitDoesNotPanic(t *testing.T) {
	fakeTelegram(t, http.StatusOK, `{"ok":true}`)
	m := newTestMessenger()
	done := make(chan struct{})
	go func() { m.Run(); close(done) }()

	m.Notify("before quit")
	close(m.Quit)
	<-done

	// must not panic (send on closed channel) after Run returned
	m.Notify("after quit")
	m.NotifyError("after quit too")
}

func TestMessengerUnconfiguredDropsWithLog(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	m := newTestMessenger()
	m.BotToken = ""
	m.send(&http.Client{Timeout: time.Second}, m.ChatID, "hello")
	if !bytes.Contains(buf.Bytes(), []byte("not configured")) {
		t.Fatalf("expected unconfigured drop to be logged, got: %s", buf.String())
	}
}

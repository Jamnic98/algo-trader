package monitoring

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// telegramAPIBase is a var so tests can point it at a local server.
var telegramAPIBase = "https://api.telegram.org"

// Messenger delivers messages to Telegram. Messages is the general channel;
// ErrorMessages goes to ErrorChatID (alerts/errors). When ErrorChatID is
// empty, NotifyError falls back to ChatID. All logging goes through the log
// package so it lands in logs/bot.log in prod (not just docker stdout).
type Messenger struct {
	BotToken      string
	ChatID        string
	ErrorChatID   string
	Messages      chan string
	ErrorMessages chan string
	Quit          chan struct{}
}

// Notify queues a general message (non-blocking; drops + logs when full).
func (m *Messenger) Notify(msg string) {
	m.enqueue(m.Messages, msg)
}

// NotifyError queues an error/alert message, preferring the dedicated error
// chat when configured.
func (m *Messenger) NotifyError(msg string) {
	if m.ErrorChatID != "" && m.ErrorMessages != nil {
		m.enqueue(m.ErrorMessages, msg)
		return
	}
	m.enqueue(m.Messages, msg)
}

func (m *Messenger) enqueue(ch chan string, msg string) {
	if ch == nil {
		return
	}
	select {
	case ch <- msg:
	default:
		log.Printf("Telegram message dropped (queue full): %s", msg)
	}
}

// Run sends queued messages until Quit is closed. It never closes the message
// channels: Notify may still be called during shutdown, and a send on a
// closed channel would panic.
func (m *Messenger) Run() {
	client := &http.Client{Timeout: 10 * time.Second}
	for {
		select {
		case msg := <-m.Messages:
			m.send(client, m.ChatID, msg)
		case msg := <-m.ErrorMessages:
			m.send(client, m.errorChat(), msg)
		case <-m.Quit:
			return
		}
	}
}

func (m *Messenger) errorChat() string {
	if m.ErrorChatID != "" {
		return m.ErrorChatID
	}
	return m.ChatID
}

// send posts one message and reports failures (bad status or ok:false) via
// the log, so a misconfigured token/chat never fails silently.
func (m *Messenger) send(client *http.Client, chatID, msg string) {
	if m.BotToken == "" || chatID == "" {
		log.Printf("Telegram not configured (BOT_TOKEN/CHAT_ID missing); message dropped: %s", msg)
		return
	}
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, m.BotToken)
	resp, err := client.PostForm(apiURL, url.Values{
		"chat_id": {chatID},
		"text":    {msg},
	})
	if err != nil {
		log.Printf("Failed to send Telegram message to chat %s: %v", chatID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		log.Printf("Telegram send to chat %s failed: status %s, body: %s", chatID, resp.Status, body)
		return
	}
	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && !result.OK {
		log.Printf("Telegram send to chat %s rejected: %s", chatID, result.Description)
	}
}

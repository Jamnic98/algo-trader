package monitoring

import (
	"fmt"
	"net/http"
	"net/url"
)

type Messenger struct {
	BotToken string
	ChatID   string
	Messages chan string
	Quit     chan struct{}
}

func (m *Messenger) Notify(msg string) {
	select {
	case m.Messages <- msg:
	default:
		fmt.Println("Telegram message dropped:", msg)
	}
}

func (m *Messenger) Run() {
	for {
		select {
		case msg := <-m.Messages:
			m.send(msg)
		case <-m.Quit:
			close(m.Messages)
			return
		}
	}
}

func (m *Messenger) send(msg string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", m.BotToken)
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id": {m.ChatID},
		"text":    {msg},
	})
	if err != nil {
		fmt.Println("Failed to send Telegram message:", err)
		return
	}
	defer resp.Body.Close()
}

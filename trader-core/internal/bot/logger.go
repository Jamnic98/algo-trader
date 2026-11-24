package bot

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type BotLogger struct {
	botID  string
	mu     sync.RWMutex
	buf    []LogEntry
	maxBuf int
	subs   []chan LogEntry
}

func NewBotLogger(botID string) *BotLogger {
	return &BotLogger{
		botID:  botID,
		maxBuf: 200,
	}
}

func (l *BotLogger) log(level, msg string) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}

	l.mu.Lock()
	// append to ring buffer
	l.buf = append(l.buf, entry)
	if len(l.buf) > l.maxBuf {
		l.buf = l.buf[len(l.buf)-l.maxBuf:]
	}
	// fan out to SSE subscribers
	for _, ch := range l.subs {
		select {
		case ch <- entry:
		default: // drop if subscriber is slow
		}
	}
	l.mu.Unlock()

	// also write to global logger
	log.Printf("[bot:%s] [%s] %s", l.botID, level, msg)
}

func (l *BotLogger) Info(format string, args ...any) {
	l.log("INFO", fmt.Sprintf(format, args...))
}

func (l *BotLogger) Warn(format string, args ...any) {
	l.log("WARN", fmt.Sprintf(format, args...))
}

func (l *BotLogger) Error(format string, args ...any) {
	l.log("ERROR", fmt.Sprintf(format, args...))
}

// Subscribe returns a channel of live log entries and a snapshot
// of the recent buffer for catch-up on connect.
func (l *BotLogger) Subscribe() ([]LogEntry, chan LogEntry) {
	ch := make(chan LogEntry, 50)
	l.mu.Lock()

	// only send entries from the last 30 minutes as snapshot
	cutoff := time.Now().Add(-30 * time.Minute)
	var snapshot []LogEntry
	for _, e := range l.buf {
		if e.Time.After(cutoff) {
			snapshot = append(snapshot, e)
		}
	}

	l.subs = append(l.subs, ch)
	l.mu.Unlock()
	return snapshot, ch
}

func (l *BotLogger) Unsubscribe(ch chan LogEntry) {
	l.mu.Lock()
	for i, s := range l.subs {
		if s == ch {
			l.subs = append(l.subs[:i], l.subs[i+1:]...)
			break
		}
	}
	l.mu.Unlock()
	close(ch)
}

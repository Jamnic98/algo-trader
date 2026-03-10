package bot

import "sync"

type Broadcaster[T any] struct {
	mu          sync.RWMutex
	subscribers []chan T
}

func (b *Broadcaster[T]) Publish(v T) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- v:
		default:
		}
	}
}

func (b *Broadcaster[T]) Subscribe() chan T {
	ch := make(chan T, 50)
	b.mu.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster[T]) Unsubscribe(ch chan T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, s := range b.subscribers {
		if s == ch {
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			close(ch)
			return
		}
	}
}

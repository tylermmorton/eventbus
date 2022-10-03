package eventbus

import (
	"errors"
	"sync"
)

type EventBus[K comparable, V any] interface {
	Subscribe(id K, ch chan V)
	Unsubscribe(id K, ch chan V) error

	Dispatch(id K, val V)
}

type bus[K comparable, V any] struct {
	mu          sync.Mutex
	subscribers map[K][]chan V
}

func New[K comparable, V any]() EventBus[K, V] {
	return &bus[K, V]{
		mu:          sync.Mutex{},
		subscribers: make(map[K][]chan V),
	}
}

func (b *bus[K, V]) Subscribe(id K, ch chan V) {
	b.mu.Lock()
	defer b.mu.Unlock()

	_, ok := b.subscribers[id]
	if !ok {
		b.subscribers[id] = make([]chan V, 0)
	}

	b.subscribers[id] = append(b.subscribers[id], ch)
}

func (b *bus[K, V]) Unsubscribe(id K, ch chan V) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	_, ok := b.subscribers[id]
	if !ok {
		return errors.New("no subscribers found for passed id")
	}

	for i := 0; i < len(b.subscribers[id]); i++ {
		if b.subscribers[id][i] == ch {
			b.subscribers[id] = append(
				b.subscribers[id][:i],
				b.subscribers[id][i+1:]...,
			)
			break
		}
	}

	return nil
}

func (b *bus[K, V]) Dispatch(id K, val V) {
	subs, ok := b.subscribers[id]
	if !ok {
		return
	}

	for _, sub := range subs {
		sub <- val
	}
}

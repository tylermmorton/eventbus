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

// Transformer takes type T1 and converts it to type T2
type Transformer[T1 any, T2 any] func(T1) (T2, error)

// Pipe creates a 'pipe' between two different EventBus instances using
// the given Transformer function. Pipe returns the closer channel that
// can be used to close the connection between the two buses.
func Pipe[
	K1 comparable, V1 any, // K1 and V1 are the key and value for the fromBus
	K2 comparable, V2 any, // K2 and V2 are the key and value for the toBus
](
	fromKey K1, fromBus EventBus[K1, V1],
	toKey K2, toBus EventBus[K2, V2],
	transformer Transformer[V1, V2],
) chan error {
	ch := make(chan V1)
	cl := make(chan error)

	fromBus.Subscribe(fromKey, ch)
	go func() {
		defer close(ch)
		<-cl

		fromBus.Unsubscribe(fromKey, ch)
	}()

	go func() {
		for {
			p, ok := <-ch
			if !ok {
				break
			}

			t, err := transformer(p)
			if err != nil {
				cl <- err

				break
			}

			toBus.Dispatch(toKey, t)
		}
	}()

	return cl
}

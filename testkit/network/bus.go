package network

import (
	"sync"
	"time"
)

// Event is a message passed between actors in the system.
type Event struct {
	Type      string
	Source    string
	Payload   any
	Timestamp time.Time
}

// Bus is an in-process channel-based message bus.
// Each subscriber (actor) has a mailbox (channel).
// Publish delivers to all non-partitioned, non-delayed subscribers.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string]chan Event
	partitions  map[string]bool
	delays      map[string]time.Duration
	recorder    *Recorder
}

// NewBus creates a message bus with an attached recorder.
func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string]chan Event),
		partitions:  make(map[string]bool),
		delays:      make(map[string]time.Duration),
		recorder:    NewRecorder(),
	}
}

// Subscribe registers a named actor and returns its mailbox.
func (b *Bus) Subscribe(name string) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, 64)
	b.subscribers[name] = ch
	return ch
}

// Publish sends an event to all non-partitioned subscribers.
func (b *Bus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for name, ch := range b.subscribers {
		if b.partitions[name] {
			continue
		}

		delay := b.delays[name]
		e := event

		if delay > 0 {
			go func(ch chan Event, e Event, d time.Duration) {
				time.Sleep(d)
				select {
				case ch <- e:
				default:
				}
			}(ch, e, delay)
			b.recorder.record(name, e)
			continue
		}

		select {
		case ch <- e:
			b.recorder.record(name, e)
		default:
		}
	}
}

// Partition isolates a subscriber -- their mailbox stops receiving events.
func (b *Bus) Partition(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.partitions[name] = true
}

// Heal restores a subscriber's connection to the bus.
func (b *Bus) Heal(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.partitions, name)
}

// Delay adds latency to a subscriber's message delivery.
func (b *Bus) Delay(name string, d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delays[name] = d
}

// ClearDelay removes latency injection for a subscriber.
func (b *Bus) ClearDelay(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.delays, name)
}

// Recorder returns the bus's event recorder.
func (b *Bus) Recorder() *Recorder {
	return b.recorder
}

package network

import (
	"fmt"
	"slices"
	"sync"
	"testing"
)

// Recorder captures events delivered to subscribers for test assertions.
type Recorder struct {
	mu     sync.Mutex
	events map[string][]Event
}

// NewRecorder creates an empty recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		events: make(map[string][]Event),
	}
}

func (r *Recorder) record(subscriber string, event Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[subscriber] = append(r.events[subscriber], event)
}

// Events returns all events recorded for a subscriber.
func (r *Recorder) Events(subscriber string) []Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Event{}, r.events[subscriber]...)
}

// AssertReceived checks that a subscriber received an event of the given type.
func (r *Recorder) AssertReceived(t testing.TB, subscriber, eventType string) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()

	if !slices.ContainsFunc(r.events[subscriber], func(e Event) bool { return e.Type == eventType }) {
		t.Errorf("subscriber %q never received event type %q", subscriber, eventType)
	}
}

// AssertNotReceived checks that a subscriber did NOT receive an event of the given type.
func (r *Recorder) AssertNotReceived(t testing.TB, subscriber, eventType string) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()

	if slices.ContainsFunc(r.events[subscriber], func(e Event) bool { return e.Type == eventType }) {
		t.Errorf("subscriber %q unexpectedly received event type %q", subscriber, eventType)
	}
}

// AssertCount checks the total number of events a subscriber received.
func (r *Recorder) AssertCount(t testing.TB, subscriber string, want int) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()

	got := len(r.events[subscriber])
	if got != want {
		t.Errorf("subscriber %q event count = %d, want %d", subscriber, got, want)
	}
}

// String returns a debug representation of recorded events.
func (r *Recorder) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var s string
	for sub, events := range r.events {
		for _, e := range events {
			s += fmt.Sprintf("[%s] %s from %s\n", sub, e.Type, e.Source)
		}
	}
	return s
}

package eventbus

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSubscribePublishUnsubscribe(t *testing.T) {
	b := New()
	defer b.Close()

	ch, unsub := b.Subscribe(EventIncidentCreated)
	if _, unsub2 := b.Subscribe(EventRCACompleted); true {
		defer unsub2()
	}

	b.Publish(Event{Type: EventIncidentCreated})
	b.Publish(Event{Type: EventRCACompleted})
	b.Publish(Event{Type: EventChaosStarted})

	got := atomic.Int32{}
	deadline := time.After(2 * time.Second)
	for {
		select {
		case e := <-ch:
			if e.Type == EventIncidentCreated {
				got.Add(1)
			}
		case <-deadline:
			t.Fatalf("timed out; received %d incident events", got.Load())
		default:
			if got.Load() == 1 {
				unsub()
				// after unsub, publishing must not panic or be received
				b.Publish(Event{Type: EventIncidentCreated})
				return
			}
		}
	}
}

func TestSubscriberTypeFiltering(t *testing.T) {
	b := New()
	defer b.Close()

	ch, unsub := b.Subscribe(EventRemediationCompleted)
	defer unsub()

	b.Publish(Event{Type: EventIncidentCreated})

	select {
	case e := <-ch:
		t.Fatalf("filtered subscriber should not receive %s", e.Type)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestToBufferedOverflowDoesNotBlock(t *testing.T) {
	b := New()
	defer b.Close()

	ch, _ := b.Subscribe(EventChaosStarted)
	// Nobody drains; 100 publishes must not block (buffer 64, drop policy).
	for i := 0; i < 100; i++ {
		b.Publish(Event{Type: EventChaosStarted})
	}
	// Drain one to prove some were buffered.
	select {
	case <-ch:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected at least one buffered event")
	}
	_ = ch
}

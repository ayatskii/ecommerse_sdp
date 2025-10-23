package observer

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockObserver struct {
	name        string
	notifyCount atomic.Int32
	lastEvent   Event
}

func (m *mockObserver) Notify(ctx context.Context, event Event) error {
	m.notifyCount.Add(1)
	m.lastEvent = event
	return nil
}

func (m *mockObserver) GetName() string {
	return m.name
}

func TestObserverPattern(t *testing.T) {
	t.Run("Attach and Notify Observers", func(t *testing.T) {
		subject := NewSubject()
		observer1 := &mockObserver{name: "observer1"}
		observer2 := &mockObserver{name: "observer2"}

		subject.Attach(observer1)
		subject.Attach(observer2)

		event := Event{
			Type:          EventPaymentSuccess,
			TransactionID: "tx-123",
			Amount:        100.00,
			Timestamp:     time.Now().Format(time.RFC3339),
		}

		subject.Notify(context.Background(), event)

		time.Sleep(100 * time.Millisecond)

		assert.Equal(t, int32(1), observer1.notifyCount.Load())
		assert.Equal(t, int32(1), observer2.notifyCount.Load())
		assert.Equal(t, EventPaymentSuccess, observer1.lastEvent.Type)
		assert.Equal(t, EventPaymentSuccess, observer2.lastEvent.Type)
	})

	t.Run("Detach Observer", func(t *testing.T) {
		subject := NewSubject()
		observer1 := &mockObserver{name: "observer1"}
		observer2 := &mockObserver{name: "observer2"}

		subject.Attach(observer1)
		subject.Attach(observer2)
		subject.Detach(observer1)

		event := Event{
			Type:      EventPaymentSuccess,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		subject.Notify(context.Background(), event)
		time.Sleep(100 * time.Millisecond)

		assert.Equal(t, int32(0), observer1.notifyCount.Load())
		assert.Equal(t, int32(1), observer2.notifyCount.Load())
	})
}

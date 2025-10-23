package observer

import (
	"context"
	"sync"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type EventType string

const (
	EventPaymentStarted EventType = "payment_started"
	EventPaymentSuccess EventType = "payment_success"
	EventPaymentFailed  EventType = "payment_failed"
	EventRefundIssued   EventType = "refund_issued"
)

type Event struct {
	Type          EventType              `json:"type"`
	TransactionID string                 `json:"transaction_id"`
	CustomerID    string                 `json:"customer_id"`
	Amount        float64                `json:"amount"`
	PaymentMethod string                 `json:"payment_method"`
	Result        *payment.PaymentResult `json:"result,omitempty"`
	Error         error                  `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     string                 `json:"timestamp"`
}

type Observer interface {
	Notify(ctx context.Context, event Event) error
	GetName() string
}

type Subject struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewSubject() *Subject {
	return &Subject{
		observers: make([]Observer, 0),
	}
}

func (s *Subject) Attach(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.observers = append(s.observers, observer)
	logger.Info("Observer attached",
		zap.String("observer", observer.GetName()),
		zap.Int("total_observers", len(s.observers)),
	)
}

func (s *Subject) Detach(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, obs := range s.observers {
		if obs.GetName() == observer.GetName() {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			logger.Info("Observer detached",
				zap.String("observer", observer.GetName()),
				zap.Int("total_observers", len(s.observers)),
			)
			return
		}
	}
}

func (s *Subject) Notify(ctx context.Context, event Event) {
	s.mu.RLock()
	observers := make([]Observer, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	logger.Info("Notifying observers",
		zap.String("event_type", string(event.Type)),
		zap.Int("observer_count", len(observers)),
	)

	var wg sync.WaitGroup
	for _, observer := range observers {
		wg.Add(1)
		go func(obs Observer) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Observer panic recovered",
						zap.String("observer", obs.GetName()),
						zap.Any("panic", r),
					)
				}
			}()

			if err := obs.Notify(ctx, event); err != nil {

				logger.Error("Observer notification failed",
					zap.String("observer", obs.GetName()),
					zap.Error(err),
				)
			} else {
				logger.Debug("Observer notified successfully",
					zap.String("observer", obs.GetName()),
					zap.String("event_type", string(event.Type)),
				)
			}
		}(observer)
	}

	wg.Wait()

	logger.Info("All observers notified",
		zap.String("event_type", string(event.Type)),
	)
}

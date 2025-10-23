package observer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type SMSNotifier struct {
	provider     string
	rateLimit    int
	messageTimes []time.Time
	mu           sync.Mutex
}

func NewSMSNotifier(provider string, rateLimit int) *SMSNotifier {
	return &SMSNotifier{
		provider:     provider,
		rateLimit:    rateLimit,
		messageTimes: make([]time.Time, 0),
	}
}

func (n *SMSNotifier) Notify(ctx context.Context, event Event) error {
	logger.Info("Sending SMS notification",
		zap.String("event_type", string(event.Type)),
		zap.String("transaction_id", event.TransactionID),
	)

	if err := n.checkRateLimit(); err != nil {
		return err
	}

	message := n.createSMSMessage(event)

	if err := n.sendSMS(ctx, message); err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	n.recordMessage()

	logger.Info("SMS sent successfully",
		zap.String("transaction_id", event.TransactionID),
	)

	return nil
}

func (n *SMSNotifier) GetName() string {
	return "sms_notifier"
}

func (n *SMSNotifier) checkRateLimit() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Minute)
	recent := []time.Time{}
	for _, t := range n.messageTimes {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	n.messageTimes = recent

	if len(n.messageTimes) >= n.rateLimit {
		return fmt.Errorf("SMS rate limit exceeded (%d messages per minute)", n.rateLimit)
	}

	return nil
}

func (n *SMSNotifier) recordMessage() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.messageTimes = append(n.messageTimes, time.Now())
}

func (n *SMSNotifier) createSMSMessage(event Event) string {
	switch event.Type {
	case EventPaymentStarted:
		return fmt.Sprintf("Payment of $%.2f is being processed. TX: %s",
			event.Amount, event.TransactionID[:8])

	case EventPaymentSuccess:
		return fmt.Sprintf("Payment of $%.2f successful! TX: %s",
			event.Amount, event.TransactionID[:8])

	case EventPaymentFailed:
		return fmt.Sprintf("Payment of $%.2f failed. TX: %s. Please try again.",
			event.Amount, event.TransactionID[:8])

	case EventRefundIssued:
		return fmt.Sprintf("Refund of $%.2f issued. TX: %s",
			event.Amount, event.TransactionID[:8])

	default:
		return fmt.Sprintf("Payment notification. TX: %s", event.TransactionID[:8])
	}
}

func (n *SMSNotifier) sendSMS(ctx context.Context, message string) error {

	time.Sleep(30 * time.Millisecond)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	logger.Debug("SMS sent",
		zap.String("provider", n.provider),
		zap.String("message", message),
	)

	return nil
}

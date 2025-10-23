package observer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type EmailNotifier struct {
	fromAddress    string
	smtpHost       string
	smtpPort       int
	workerPoolSize int
	emailQueue     chan EmailMessage
	wg             sync.WaitGroup
	started        bool
	mu             sync.Mutex
}

type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

func NewEmailNotifier(fromAddress, smtpHost string, smtpPort, workerPoolSize int) *EmailNotifier {
	notifier := &EmailNotifier{
		fromAddress:    fromAddress,
		smtpHost:       smtpHost,
		smtpPort:       smtpPort,
		workerPoolSize: workerPoolSize,
		emailQueue:     make(chan EmailMessage, 100),
	}

	notifier.startWorkers()
	return notifier
}

func (n *EmailNotifier) startWorkers() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.started {
		return
	}

	for i := 0; i < n.workerPoolSize; i++ {
		n.wg.Add(1)
		go n.worker(i)
	}

	n.started = true
	logger.Info("Email worker pool started",
		zap.Int("workers", n.workerPoolSize),
	)
}

func (n *EmailNotifier) worker(id int) {
	defer n.wg.Done()

	logger.Info("Email worker started",
		zap.Int("worker_id", id),
	)

	for msg := range n.emailQueue {
		if err := n.sendEmail(msg); err != nil {
			logger.Error("Failed to send email",
				zap.Int("worker_id", id),
				zap.String("to", msg.To),
				zap.Error(err),
			)
		} else {
			logger.Info("Email sent successfully",
				zap.Int("worker_id", id),
				zap.String("to", msg.To),
			)
		}
	}

	logger.Info("Email worker stopped",
		zap.Int("worker_id", id),
	)
}

func (n *EmailNotifier) Notify(ctx context.Context, event Event) error {
	logger.Info("Queueing email notification",
		zap.String("event_type", string(event.Type)),
		zap.String("transaction_id", event.TransactionID),
	)

	msg := n.createEmailMessage(event)

	select {
	case n.emailQueue <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		logger.Warn("Email queue full, dropping message")
		return fmt.Errorf("email queue full")
	}
}

func (n *EmailNotifier) GetName() string {
	return "email_notifier"
}

func (n *EmailNotifier) createEmailMessage(event Event) EmailMessage {
	var subject, body string

	switch event.Type {
	case EventPaymentStarted:
		subject = "Payment Processing Started"
		body = fmt.Sprintf(
			"Your payment of $%.2f has been initiated.\nTransaction ID: %s",
			event.Amount, event.TransactionID,
		)

	case EventPaymentSuccess:
		subject = "Payment Successful"
		body = fmt.Sprintf(
			"Your payment of $%.2f has been processed successfully.\nTransaction ID: %s\nPayment Method: %s",
			event.Amount, event.TransactionID, event.PaymentMethod,
		)

	case EventPaymentFailed:
		subject = "Payment Failed"
		body = fmt.Sprintf(
			"Your payment of $%.2f has failed.\nTransaction ID: %s\nPlease try again or contact support.",
			event.Amount, event.TransactionID,
		)

	case EventRefundIssued:
		subject = "Refund Issued"
		body = fmt.Sprintf(
			"A refund of $%.2f has been issued to your account.\nTransaction ID: %s",
			event.Amount, event.TransactionID,
		)

	default:
		subject = "Payment Notification"
		body = fmt.Sprintf("Transaction ID: %s", event.TransactionID)
	}

	return EmailMessage{
		To:      "customer@example.com",
		Subject: subject,
		Body:    body,
	}
}

func (n *EmailNotifier) sendEmail(msg EmailMessage) error {

	time.Sleep(50 * time.Millisecond)

	logger.Debug("Email sent",
		zap.String("to", msg.To),
		zap.String("subject", msg.Subject),
	)

	return nil
}

func (n *EmailNotifier) Close() {
	n.mu.Lock()
	if !n.started {
		n.mu.Unlock()
		return
	}
	n.mu.Unlock()

	close(n.emailQueue)
	n.wg.Wait()
	logger.Info("Email notifier closed")
}

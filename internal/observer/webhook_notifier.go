package observer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type WebhookNotifier struct {
	url           string
	timeout       time.Duration
	retryAttempts int
	client        *http.Client
}

func NewWebhookNotifier(url string, timeout time.Duration, retryAttempts int) *WebhookNotifier {
	return &WebhookNotifier{
		url:           url,
		timeout:       timeout,
		retryAttempts: retryAttempts,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (n *WebhookNotifier) Notify(ctx context.Context, event Event) error {
	logger.Info("Sending webhook notification",
		zap.String("event_type", string(event.Type)),
		zap.String("transaction_id", event.TransactionID),
		zap.String("url", n.url),
	)

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= n.retryAttempts; attempt++ {
		if attempt > 0 {
			logger.Info("Retrying webhook",
				zap.Int("attempt", attempt),
				zap.String("transaction_id", event.TransactionID),
			)

			time.Sleep(time.Duration(attempt) * time.Second)
		}

		err := n.sendWebhook(ctx, payload)
		if err == nil {
			logger.Info("Webhook sent successfully",
				zap.String("transaction_id", event.TransactionID),
				zap.Int("attempts", attempt+1),
			)
			return nil
		}

		lastErr = err
		logger.Warn("Webhook attempt failed",
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", n.retryAttempts+1, lastErr)
}

func (n *WebhookNotifier) GetName() string {
	return "webhook_notifier"
}

func (n *WebhookNotifier) sendWebhook(ctx context.Context, payload []byte) error {

	req, err := http.NewRequestWithContext(ctx, "POST", n.url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ECommerce-Payment-System/1.0")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}

package observer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type AuditLogger struct {
	logPath string
	file    *os.File
	mu      sync.Mutex
}

func NewAuditLogger(logPath string) (*AuditLogger, error) {

	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}

	return &AuditLogger{
		logPath: logPath,
		file:    file,
	}, nil
}

func (a *AuditLogger) Notify(ctx context.Context, event Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	logger.Info("Writing to audit log",
		zap.String("event_type", string(event.Type)),
		zap.String("transaction_id", event.TransactionID),
	)

	entry := AuditEntry{
		Timestamp:     time.Now().Format(time.RFC3339),
		EventType:     string(event.Type),
		TransactionID: event.TransactionID,
		CustomerID:    event.CustomerID,
		Amount:        event.Amount,
		PaymentMethod: event.PaymentMethod,
		Metadata:      event.Metadata,
	}

	if event.Error != nil {
		entry.Error = event.Error.Error()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	if _, err := a.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	if err := a.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync audit log: %w", err)
	}

	logger.Debug("Audit entry written",
		zap.String("transaction_id", event.TransactionID),
	)

	return nil
}

func (a *AuditLogger) GetName() string {
	return "audit_logger"
}

func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file != nil {
		return a.file.Close()
	}
	return nil
}

type AuditEntry struct {
	Timestamp     string                 `json:"timestamp"`
	EventType     string                 `json:"event_type"`
	TransactionID string                 `json:"transaction_id"`
	CustomerID    string                 `json:"customer_id"`
	Amount        float64                `json:"amount"`
	PaymentMethod string                 `json:"payment_method"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

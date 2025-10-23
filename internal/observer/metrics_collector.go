package observer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type MetricsCollector struct {
	successCount   atomic.Int64
	failureCount   atomic.Int64
	totalAmount    atomic.Uint64
	paymentCounts  map[string]*atomic.Int64
	lastExport     time.Time
	exportInterval time.Duration
	mu             sync.RWMutex
}

func NewMetricsCollector(exportInterval time.Duration) *MetricsCollector {
	return &MetricsCollector{
		paymentCounts:  make(map[string]*atomic.Int64),
		exportInterval: exportInterval,
		lastExport:     time.Now(),
	}
}

func (m *MetricsCollector) Notify(ctx context.Context, event Event) error {
	logger.Debug("Collecting metrics",
		zap.String("event_type", string(event.Type)),
		zap.String("transaction_id", event.TransactionID),
	)

	switch event.Type {
	case EventPaymentSuccess:
		m.successCount.Add(1)
		m.addAmount(event.Amount)
		m.incrementPaymentMethodCount(event.PaymentMethod)

	case EventPaymentFailed:
		m.failureCount.Add(1)

	case EventRefundIssued:
		m.addAmount(-event.Amount)
	}

	m.maybeExportMetrics()

	return nil
}

func (m *MetricsCollector) GetName() string {
	return "metrics_collector"
}

func (m *MetricsCollector) addAmount(amount float64) {
	cents := uint64(amount * 100)
	m.totalAmount.Add(cents)
}

func (m *MetricsCollector) incrementPaymentMethodCount(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	counter, exists := m.paymentCounts[method]
	if !exists {
		counter = &atomic.Int64{}
		m.paymentCounts[method] = counter
	}
	counter.Add(1)
}

func (m *MetricsCollector) maybeExportMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if time.Since(m.lastExport) < m.exportInterval {
		return
	}

	m.exportMetrics()
	m.lastExport = time.Now()
}

func (m *MetricsCollector) exportMetrics() {
	successCount := m.successCount.Load()
	failureCount := m.failureCount.Load()
	totalAmount := float64(m.totalAmount.Load()) / 100.0

	totalCount := successCount + failureCount
	successRate := 0.0
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount) * 100.0
	}

	logger.Info("Payment Metrics",
		zap.Int64("total_payments", totalCount),
		zap.Int64("successful_payments", successCount),
		zap.Int64("failed_payments", failureCount),
		zap.Float64("success_rate", successRate),
		zap.Float64("total_amount", totalAmount),
	)

	for method, counter := range m.paymentCounts {
		count := counter.Load()
		logger.Info("Payment Method Metrics",
			zap.String("method", method),
			zap.Int64("count", count),
		)
	}
}

func (m *MetricsCollector) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	paymentMethodCounts := make(map[string]int64)
	for method, counter := range m.paymentCounts {
		paymentMethodCounts[method] = counter.Load()
	}

	return Metrics{
		SuccessCount:        m.successCount.Load(),
		FailureCount:        m.failureCount.Load(),
		TotalAmount:         float64(m.totalAmount.Load()) / 100.0,
		PaymentMethodCounts: paymentMethodCounts,
	}
}

func (m *MetricsCollector) Reset() {
	m.successCount.Store(0)
	m.failureCount.Store(0)
	m.totalAmount.Store(0)

	m.mu.Lock()
	m.paymentCounts = make(map[string]*atomic.Int64)
	m.mu.Unlock()

	logger.Info("Metrics reset")
}

type Metrics struct {
	SuccessCount        int64            `json:"success_count"`
	FailureCount        int64            `json:"failure_count"`
	TotalAmount         float64          `json:"total_amount"`
	PaymentMethodCounts map[string]int64 `json:"payment_method_counts"`
}

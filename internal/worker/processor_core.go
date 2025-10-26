package worker

import (
    "context"
    "log"
    "time"

    "CS6650_Online_Store/internal/models"
)

// processOrder performs the core order processing (payment simulation).
// It is cancelable via context and uses the processor's paymentGateway
// to simulate the external payment bottleneck (1 at a time).
func (p *OrderProcessor) processOrder(ctx context.Context, order *models.Order) error {
    start := time.Now()

    // Try to acquire the payment gateway token or return if context canceled
    select {
    case p.paymentGateway <- struct{}{}:
        // acquired
    case <-ctx.Done():
        return ctx.Err()
    }

    // Ensure we always release the token
    defer func() {
        <-p.paymentGateway
    }()

    // Simulate payment verification, but respect ctx cancellation
    select {
    case <-time.After(3 * time.Second):
        // processed
    case <-ctx.Done():
        return ctx.Err()
    }

    order.Status = models.StatusCompleted
    log.Printf("Order %s payment completed in %v", order.OrderID, time.Since(start))
    return nil
}

// NewLocalProcessor returns an OrderProcessor suitable for local/invoked
// contexts (like Lambda) where only the paymentGateway serialization is
// required. The returned processor will not have an SQS client configured.
func NewLocalProcessor() *OrderProcessor {
    return &OrderProcessor{
        paymentGateway: make(chan struct{}, 1),
        shutdown:       make(chan struct{}),
    }
}

// Process is the exported wrapper around the internal processOrder method.
// It allows callers from other packages (Lambda handler) to invoke the
// payment processing logic while keeping the internal implementation private.
func (p *OrderProcessor) Process(ctx context.Context, order *models.Order) error {
    return p.processOrder(ctx, order)
}

// Package ai contains replaceable implementations of stable core contracts.
package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/kntjspr/fullstack-golang-next-template/core"
)

// PaymentProcessor is the AI-owned implementation of core.PaymentProcessor.
type PaymentProcessor struct{}

// NewPaymentProcessor constructs a payment processor implementation.
func NewPaymentProcessor() core.PaymentProcessor {
	return PaymentProcessor{}
}

// Authorize validates an amount and returns a transaction ID for a successful authorization.
func (PaymentProcessor) Authorize(ctx context.Context, amount int64) (core.TransactionID, error) {
	if err := core.ValidateAuthorization(ctx, amount); err != nil {
		return "", err
	}

	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}

	return core.TransactionID("txn_" + hex.EncodeToString(id)), nil
}

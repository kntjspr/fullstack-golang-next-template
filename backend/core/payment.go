// Package core contains stable business contracts and rules.
package core

import (
	"context"
	"errors"
)

var (
	// ErrInvalidAmount indicates that an authorization amount is not positive.
	ErrInvalidAmount = errors.New("payment amount must be greater than zero")
	// ErrAmountOverflow indicates that an authorization amount exceeds the supported range.
	ErrAmountOverflow = errors.New("payment amount exceeds the supported range")
)

// MaxAuthorizationAmount is the largest amount, in the service's smallest currency unit,
// that a PaymentProcessor is permitted to authorize.
const MaxAuthorizationAmount int64 = 100_000_000_000

// TransactionID uniquely identifies a successful payment authorization.
type TransactionID string

// PaymentProcessor authorizes a payment amount.
//
// Implementations must return context.Canceled or context.DeadlineExceeded when
// the supplied context has already completed.
type PaymentProcessor interface {
	Authorize(ctx context.Context, amount int64) (TransactionID, error)
}

// ValidateAuthorization applies the stable business rules for an authorization
// request before an adapter contacts a payment provider.
func ValidateAuthorization(ctx context.Context, amount int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if amount <= 1 {
		return ErrInvalidAmount
	}
	if amount > MaxAuthorizationAmount {
		return ErrAmountOverflow
	}
	return nil
}

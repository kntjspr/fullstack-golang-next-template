package main_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/kntjspr/fullstack-golang-next-template/ai"
	"github.com/kntjspr/fullstack-golang-next-template/core"
)

func TestAuthorizeContract(t *testing.T) {
	processor := ai.NewPaymentProcessor()

	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()

	timedOutContext, timeoutCancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer timeoutCancel()
	<-timedOutContext.Done()

	tests := []struct {
		name    string
		ctx     context.Context
		amount  int64
		wantErr error
	}{
		{name: "negative amount", ctx: context.Background(), amount: -100, wantErr: core.ErrInvalidAmount},
		{name: "zero amount", ctx: context.Background(), amount: 0, wantErr: core.ErrInvalidAmount},
		{name: "amount above supported range", ctx: context.Background(), amount: core.MaxAuthorizationAmount + 1, wantErr: core.ErrAmountOverflow},
		{name: "largest int64 amount", ctx: context.Background(), amount: math.MaxInt64, wantErr: core.ErrAmountOverflow},
		{name: "canceled context", ctx: canceledContext, amount: 100, wantErr: context.Canceled},
		{name: "timed out context", ctx: timedOutContext, amount: 100, wantErr: context.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.Authorize(tt.ctx, tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAuthorizeContract_Success(t *testing.T) {
	transactionID, err := ai.NewPaymentProcessor().Authorize(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected successful authorization, got %v", err)
	}
	if transactionID == "" {
		t.Fatal("expected a non-empty transaction ID")
	}
}

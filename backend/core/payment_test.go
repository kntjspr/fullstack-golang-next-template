package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestValidateAuthorization(t *testing.T) {
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
		{name: "negative amount", ctx: context.Background(), amount: -1, wantErr: ErrInvalidAmount},
		{name: "zero amount", ctx: context.Background(), amount: 0, wantErr: ErrInvalidAmount},
		{name: "amount above supported range", ctx: context.Background(), amount: MaxAuthorizationAmount + 1, wantErr: ErrAmountOverflow},
		{name: "canceled context", ctx: canceledContext, amount: 1, wantErr: context.Canceled},
		{name: "timed out context", ctx: timedOutContext, amount: 1, wantErr: context.DeadlineExceeded},
		{name: "smallest valid amount", ctx: context.Background(), amount: 1},
		{name: "largest valid amount", ctx: context.Background(), amount: MaxAuthorizationAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthorization(tt.ctx, tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

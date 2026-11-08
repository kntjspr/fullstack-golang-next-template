package main_test

import (
	"testing"

	"myproject2/ai"
)

func TestNewPaymentProcessor(t *testing.T) {
	if ai.NewPaymentProcessor() == nil {
		t.Fatal("expected a payment processor")
	}
}

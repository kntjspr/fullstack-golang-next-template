package main_test

import (
	"testing"

	"github.com/kntjspr/fullstack-golang-next-template/ai"
)

func TestNewPaymentProcessor(t *testing.T) {
	if ai.NewPaymentProcessor() == nil {
		t.Fatal("expected a payment processor")
	}
}

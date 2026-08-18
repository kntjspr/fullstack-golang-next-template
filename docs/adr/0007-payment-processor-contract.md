# ADR 0007: Define a Stable Payment Processor Boundary

- Status: Accepted

## Context

Payment authorization is a business capability that may need provider adapters, including AI-authored adapters. Provider details must not redefine the behavior clients depend on.

## Decision

Define `core.PaymentProcessor` as the stable boundary. Its amount and context rules are implemented in `core.ValidateAuthorization`; adapters in `ai/` delegate to those rules and supply provider-specific behavior. `service_test.go` is the hand-authored contract suite for the boundary.

## Consequences

- Payment adapters can be replaced without changing callers.
- Negative, zero, out-of-range, cancelled, and timed-out requests have stable outcomes.
- A successful authorization returns a non-empty transaction ID.
- Changes to established contract assertions are rejected by CI; new cases remain additive.

package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// Redis defines the Redis behavior consumed by application services.
type Redis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Ping(ctx context.Context) error
}

// RedisMock is a testify-backed mock implementation of Redis.
type RedisMock struct {
	mock.Mock
}

func (m *RedisMock) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *RedisMock) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *RedisMock) Del(ctx context.Context, keys ...string) (int64, error) {
	callArgs := []any{ctx}
	for _, key := range keys {
		callArgs = append(callArgs, key)
	}

	args := m.Called(callArgs...)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisMock) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

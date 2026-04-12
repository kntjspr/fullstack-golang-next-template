package config

import (
	"strings"
	"sync"
	"testing"
)

func TestNewConfig_DefaultServerPortWhenMissing(t *testing.T) {
	resetConfigSingletonForTest()
	t.Cleanup(resetConfigSingletonForTest)

	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "")
	t.Setenv("SERVER_READ_TIMEOUT", "5")
	t.Setenv("SERVER_WRITE_TIMEOUT", "10")
	t.Setenv("SERVER_IDLE_TIMEOUT", "120")

	cfg := mustNotPanic(t, NewConfig)
	if cfg.Server.Addr != "127.0.0.1:5000" {
		t.Fatalf("unexpected server addr: got %q want %q", cfg.Server.Addr, "127.0.0.1:5000")
	}
}

func TestNewConfig_PanicsOnInvalidServerPort(t *testing.T) {
	resetConfigSingletonForTest()
	t.Cleanup(resetConfigSingletonForTest)

	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "not-a-number")
	t.Setenv("SERVER_READ_TIMEOUT", "5")
	t.Setenv("SERVER_WRITE_TIMEOUT", "10")
	t.Setenv("SERVER_IDLE_TIMEOUT", "120")

	panicValue := mustPanic(t, NewConfig)
	panicMessage, ok := panicValue.(string)
	if !ok {
		t.Fatalf("panic value type: got %T want string", panicValue)
	}
	if !strings.Contains(panicMessage, "wrong server port") {
		t.Fatalf("unexpected panic message: %q", panicMessage)
	}
}

func resetConfigSingletonForTest() {
	once = sync.Once{}
	instance = nil
}

func mustNotPanic(t *testing.T, fn func() *Config) *Config {
	t.Helper()

	var (
		cfg       *Config
		panicData any
	)

	func() {
		defer func() {
			panicData = recover()
		}()

		cfg = fn()
	}()

	if panicData != nil {
		t.Fatalf("unexpected panic: %v", panicData)
	}
	if cfg == nil {
		t.Fatal("expected config instance, got nil")
	}

	return cfg
}

func mustPanic(t *testing.T, fn func() *Config) any {
	t.Helper()

	var panicData any
	func() {
		defer func() {
			panicData = recover()
		}()

		_ = fn()
	}()

	if panicData == nil {
		t.Fatal("expected panic, got nil")
	}

	return panicData
}

package config

import (
	"strings"
	"testing"
)

func TestNewConfig_DefaultServerPortWhenMissing(t *testing.T) {
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "")
	t.Setenv("SERVER_READ_TIMEOUT", "5")
	t.Setenv("SERVER_WRITE_TIMEOUT", "10")
	t.Setenv("SERVER_IDLE_TIMEOUT", "120")
	t.Setenv("STAGE_STATUS", "dev")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}
	if cfg.Server.Addr != "127.0.0.1:5000" {
		t.Fatalf("unexpected server addr: got %q want %q", cfg.Server.Addr, "127.0.0.1:5000")
	}
}

func TestNewConfig_ErrorsOnInvalidServerPort(t *testing.T) {
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "not-a-number")
	t.Setenv("SERVER_READ_TIMEOUT", "5")
	t.Setenv("SERVER_WRITE_TIMEOUT", "10")
	t.Setenv("SERVER_IDLE_TIMEOUT", "120")
	t.Setenv("STAGE_STATUS", "dev")

	_, err := NewConfig()
	if err == nil {
		t.Fatal("expected error for invalid SERVER_PORT")
	}
	if got, want := err.Error(), "SERVER_PORT"; !strings.Contains(got, want) {
		t.Fatalf("error should mention %q, got %q", want, got)
	}
}

func TestNewConfig_ErrorsOnInvalidStageStatus(t *testing.T) {
	t.Setenv("STAGE_STATUS", "production")

	_, err := NewConfig()
	if err == nil {
		t.Fatal("expected error for invalid STAGE_STATUS")
	}
}

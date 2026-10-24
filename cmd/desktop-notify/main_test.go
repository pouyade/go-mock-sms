package main

import (
	"os"
	"testing"
)

func TestParseConfigDefaults(t *testing.T) {
	t.Setenv("DESKTOP_NOTIFY_HOST", "")
	t.Setenv("DESKTOP_NOTIFY_PORT", "")
	t.Setenv("DESKTOP_NOTIFY_TITLE", "")
	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if cfg.host != defaultHost || cfg.port != defaultPort || cfg.title != "Desktop Notify" {
		t.Fatalf("config = %#v, want defaults", cfg)
	}
}

func TestParseConfigEnvironmentAndFlags(t *testing.T) {
	t.Setenv("DESKTOP_NOTIFY_HOST", "127.0.0.2")
	t.Setenv("DESKTOP_NOTIFY_PORT", "5555")
	t.Setenv("DESKTOP_NOTIFY_TITLE", "From environment")
	cfg, err := parseConfig([]string{"--host", "localhost", "--port", "5556", "--title", "From flag"})
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if cfg.host != "localhost" || cfg.port != 5556 || cfg.title != "From flag" {
		t.Fatalf("config = %#v, want flag values", cfg)
	}
}

func TestParseConfigInvalid(t *testing.T) {
	tests := [][]string{
		{"--port", "0"},
		{"--port", "65536"},
		{"--host", ""},
		{"unexpected"},
	}
	for _, args := range tests {
		if _, err := parseConfig(args); err == nil {
			t.Errorf("parseConfig(%v) returned nil error", args)
		}
	}
}

func TestEnvironmentHelpers(t *testing.T) {
	const name = "DESKTOP_NOTIFY_TEST_VALUE"
	_ = os.Unsetenv(name)
	if got := envOrDefault(name, "fallback"); got != "fallback" {
		t.Fatalf("envOrDefault() = %q, want fallback", got)
	}
	t.Setenv(name, "42")
	if got := envIntOrDefault(name, 7); got != 42 {
		t.Fatalf("envIntOrDefault() = %d, want 42", got)
	}
	t.Setenv(name, "not-a-number")
	if got := envIntOrDefault(name, 7); got != 7 {
		t.Fatalf("envIntOrDefault() = %d, want 7", got)
	}
}

func TestIsLoopbackHost(t *testing.T) {
	if !isLoopbackHost("localhost") || !isLoopbackHost("127.0.0.1") || !isLoopbackHost("::1") {
		t.Fatal("expected loopback hosts to be recognized")
	}
	if isLoopbackHost("0.0.0.0") || isLoopbackHost("example.com") {
		t.Fatal("expected non-loopback hosts to be rejected")
	}
}

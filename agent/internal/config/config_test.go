package config

import (
	"io"
	"testing"
	"time"
)

func TestLoadDryRunUsesDefaults(t *testing.T) {
	config, err := load([]string{"--dry-run"}, func(string) string { return "" }, io.Discard)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.DiskPath != "/" {
		t.Fatalf("DiskPath = %q, want /", config.DiskPath)
	}
	if config.Interval != 15*time.Second {
		t.Fatalf("Interval = %v, want 15s", config.Interval)
	}
}

func TestLoadRequiresCredentialsOutsideDryRun(t *testing.T) {
	_, err := load(nil, func(string) string { return "" }, io.Discard)
	if err == nil {
		t.Fatal("load config returned nil error, want validation error")
	}
}

func TestFlagsOverrideEnvironment(t *testing.T) {
	environment := map[string]string{
		"VPSENTINEL_SERVER_URL": "https://from-env.example/metrics",
		"VPSENTINEL_TOKEN":      "env-token",
		"VPSENTINEL_INTERVAL":   "30s",
	}
	getenv := func(key string) string { return environment[key] }
	config, err := load([]string{"--interval=5s"}, getenv, io.Discard)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Interval != 5*time.Second {
		t.Fatalf("Interval = %v, want 5s", config.Interval)
	}
}

func TestLoadRejectsPlainHTTPByDefault(t *testing.T) {
	environment := map[string]string{
		"VPSENTINEL_SERVER_URL": "http://monitor.example/metrics",
		"VPSENTINEL_TOKEN":      "secret",
	}
	_, err := load(nil, func(key string) string { return environment[key] }, io.Discard)
	if err == nil {
		t.Fatal("load config returned nil error, want insecure HTTP error")
	}
}

func TestLoadAllowsExplicitPlainHTTP(t *testing.T) {
	environment := map[string]string{
		"VPSENTINEL_SERVER_URL":          "http://127.0.0.1:8080/metrics",
		"VPSENTINEL_TOKEN":               "secret",
		"VPSENTINEL_ALLOW_INSECURE_HTTP": "true",
	}
	config, err := load(nil, func(key string) string { return environment[key] }, io.Discard)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !config.AllowInsecureHTTP {
		t.Fatal("AllowInsecureHTTP = false, want true")
	}
}

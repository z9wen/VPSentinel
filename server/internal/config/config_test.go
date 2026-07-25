package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	cfg, err := load(testEnvironment(func(key string) string {
		switch key {
		case "VPSENTINEL_DATABASE_URL":
			return "postgres://vpsentinel:test@localhost/vpsentinel"
		}
		return ""
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.ListenAddress != ":8080" {
		t.Fatalf("ListenAddress = %q, want :8080", cfg.ListenAddress)
	}
	if cfg.MaxBodyBytes != 1<<20 {
		t.Fatalf("MaxBodyBytes = %d, want %d", cfg.MaxBodyBytes, 1<<20)
	}
	if !cfg.AutoMigrate {
		t.Fatal("AutoMigrate = false, want true")
	}
}

func TestLoadRequiresStrongDevelopmentToken(t *testing.T) {
	_, err := load(testEnvironment(func(key string) string {
		switch key {
		case "VPSENTINEL_AGENT_TOKEN":
			return "short"
		case "VPSENTINEL_DATABASE_URL":
			return "postgres://vpsentinel:test@localhost/vpsentinel"
		}
		return ""
	}))
	if err == nil {
		t.Fatal("load returned nil error, want token validation error")
	}
}

func testEnvironment(override func(string) string) func(string) string {
	return func(key string) string {
		if value := override(key); value != "" {
			return value
		}
		switch key {
		case "VPSENTINEL_AGENT_TOKEN":
			return "0123456789abcdef"
		case "VPSENTINEL_ADMIN_EMAIL":
			return "admin@example.com"
		case "VPSENTINEL_ADMIN_PASSWORD":
			return "test-password-123"
		case "VPSENTINEL_JWT_SECRET":
			return "0123456789abcdef0123456789abcdef"
		case "VPSENTINEL_GRAFANA_WEBHOOK_SECRET":
			return "grafana-test-secret"
		}
		return ""
	}
}

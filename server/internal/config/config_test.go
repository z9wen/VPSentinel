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

func TestLoadParsesTrustedProxies(t *testing.T) {
	cfg, err := load(testEnvironment(func(key string) string {
		switch key {
		case "VPSENTINEL_DATABASE_URL":
			return "postgres://vpsentinel:test@localhost/vpsentinel"
		case "VPSENTINEL_TRUSTED_PROXIES":
			return " 172.16.0.0/12, 127.0.0.1 ,"
		}
		return ""
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.TrustedProxies) != 2 || cfg.TrustedProxies[0] != "172.16.0.0/12" || cfg.TrustedProxies[1] != "127.0.0.1" {
		t.Fatalf("TrustedProxies = %q, want [172.16.0.0/12 127.0.0.1]", cfg.TrustedProxies)
	}
}

func TestLoadRejectsInvalidTrustedProxy(t *testing.T) {
	_, err := load(testEnvironment(func(key string) string {
		switch key {
		case "VPSENTINEL_DATABASE_URL":
			return "postgres://vpsentinel:test@localhost/vpsentinel"
		case "VPSENTINEL_TRUSTED_PROXIES":
			return "web-proxy"
		}
		return ""
	}))
	if err == nil {
		t.Fatal("load returned nil error, want trusted proxy validation error")
	}
}

func TestLoadMigrationNeedsOnlyDatabaseSettings(t *testing.T) {
	cfg, err := loadMigration(func(key string) string {
		if key == "VPSENTINEL_DATABASE_URL" {
			return "postgres://vpsentinel:test@localhost/vpsentinel"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("loadMigration: %v", err)
	}
	if cfg.DatabaseStartTimeout != defaultDatabaseStartTimeout {
		t.Fatalf("DatabaseStartTimeout = %s, want %s", cfg.DatabaseStartTimeout, defaultDatabaseStartTimeout)
	}
	if _, err := loadMigration(func(string) string { return "" }); err == nil {
		t.Fatal("loadMigration returned nil error without a database URL")
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

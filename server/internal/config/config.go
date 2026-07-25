package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultListenAddress        = ":8080"
	defaultNodeID               = "development-node"
	defaultMaxBodyBytes         = int64(1 << 20)
	defaultAutoMigrate          = true
	defaultDatabaseStartTimeout = 30 * time.Second
	defaultShutdownTimeout      = 10 * time.Second
	defaultAccessTokenTTL       = 15 * time.Minute
	defaultRefreshTokenTTL      = 7 * 24 * time.Hour
	defaultRetentionDays        = 30
)

type Config struct {
	ListenAddress        string
	AgentToken           string
	NodeID               string
	GinMode              string
	MaxBodyBytes         int64
	DatabaseURL          string
	AutoMigrate          bool
	DatabaseStartTimeout time.Duration
	ShutdownTimeout      time.Duration
	AdminEmail           string
	AdminPassword        string
	JWTSecret            string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	WebOrigin            string
	CookieSecure         bool
	GrafanaWebhookSecret string
	GrafanaURL           string
	RetentionDays        int
}

func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	maxBodyBytes, err := int64Value(getenv("VPSENTINEL_MAX_BODY_BYTES"), defaultMaxBodyBytes)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_MAX_BODY_BYTES: %w", err)
	}
	autoMigrate, err := boolValue(getenv("VPSENTINEL_AUTO_MIGRATE"), defaultAutoMigrate)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_AUTO_MIGRATE: %w", err)
	}
	databaseStartTimeout, err := durationValue(getenv("VPSENTINEL_DATABASE_START_TIMEOUT"), defaultDatabaseStartTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_DATABASE_START_TIMEOUT: %w", err)
	}
	shutdownTimeout, err := durationValue(getenv("VPSENTINEL_SHUTDOWN_TIMEOUT"), defaultShutdownTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_SHUTDOWN_TIMEOUT: %w", err)
	}
	accessTokenTTL, err := durationValue(getenv("VPSENTINEL_ACCESS_TOKEN_TTL"), defaultAccessTokenTTL)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_ACCESS_TOKEN_TTL: %w", err)
	}
	refreshTokenTTL, err := durationValue(getenv("VPSENTINEL_REFRESH_TOKEN_TTL"), defaultRefreshTokenTTL)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_REFRESH_TOKEN_TTL: %w", err)
	}
	cookieSecure, err := boolValue(getenv("VPSENTINEL_COOKIE_SECURE"), false)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_COOKIE_SECURE: %w", err)
	}
	retentionDays, err := intValue(getenv("VPSENTINEL_RETENTION_DAYS"), defaultRetentionDays)
	if err != nil {
		return Config{}, fmt.Errorf("parse VPSENTINEL_RETENTION_DAYS: %w", err)
	}

	cfg := Config{
		ListenAddress:        valueOrDefault(getenv("VPSENTINEL_LISTEN_ADDRESS"), defaultListenAddress),
		AgentToken:           getenv("VPSENTINEL_AGENT_TOKEN"),
		NodeID:               valueOrDefault(getenv("VPSENTINEL_NODE_ID"), defaultNodeID),
		GinMode:              valueOrDefault(getenv("GIN_MODE"), gin.ReleaseMode),
		MaxBodyBytes:         maxBodyBytes,
		DatabaseURL:          getenv("VPSENTINEL_DATABASE_URL"),
		AutoMigrate:          autoMigrate,
		DatabaseStartTimeout: databaseStartTimeout,
		ShutdownTimeout:      shutdownTimeout,
		AdminEmail:           getenv("VPSENTINEL_ADMIN_EMAIL"),
		AdminPassword:        getenv("VPSENTINEL_ADMIN_PASSWORD"),
		JWTSecret:            getenv("VPSENTINEL_JWT_SECRET"),
		AccessTokenTTL:       accessTokenTTL,
		RefreshTokenTTL:      refreshTokenTTL,
		WebOrigin:            getenv("VPSENTINEL_WEB_ORIGIN"),
		CookieSecure:         cookieSecure,
		GrafanaWebhookSecret: getenv("VPSENTINEL_GRAFANA_WEBHOOK_SECRET"),
		GrafanaURL:           valueOrDefault(getenv("VPSENTINEL_GRAFANA_URL"), "/grafana"),
		RetentionDays:        retentionDays,
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.ListenAddress == "" {
		return errors.New("listen address must not be empty")
	}
	if len(c.AgentToken) < 16 {
		return errors.New("VPSENTINEL_AGENT_TOKEN must contain at least 16 characters")
	}
	if c.NodeID == "" {
		return errors.New("node ID must not be empty")
	}
	if c.MaxBodyBytes <= 0 {
		return errors.New("max body bytes must be greater than zero")
	}
	if c.DatabaseURL == "" {
		return errors.New("VPSENTINEL_DATABASE_URL is required")
	}
	if c.DatabaseStartTimeout <= 0 {
		return errors.New("database start timeout must be greater than zero")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be greater than zero")
	}
	if c.AdminEmail == "" || c.AdminPassword == "" {
		return errors.New("VPSENTINEL_ADMIN_EMAIL and VPSENTINEL_ADMIN_PASSWORD are required")
	}
	if len(c.AdminPassword) < 12 {
		return errors.New("VPSENTINEL_ADMIN_PASSWORD must contain at least 12 characters")
	}
	if len(c.JWTSecret) < 32 {
		return errors.New("VPSENTINEL_JWT_SECRET must contain at least 32 characters")
	}
	if c.AccessTokenTTL <= 0 || c.RefreshTokenTTL <= c.AccessTokenTTL {
		return errors.New("refresh token TTL must be greater than the positive access token TTL")
	}
	if len(c.GrafanaWebhookSecret) < 16 {
		return errors.New("VPSENTINEL_GRAFANA_WEBHOOK_SECRET must contain at least 16 characters")
	}
	if c.RetentionDays <= 0 {
		return errors.New("retention days must be greater than zero")
	}
	switch c.GinMode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
	default:
		return fmt.Errorf("unsupported Gin mode %q", c.GinMode)
	}
	return nil
}

func intValue(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func boolValue(value string, fallback bool) (bool, error) {
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseBool(value)
}

func int64Value(value string, fallback int64) (int64, error) {
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func durationValue(value string, fallback time.Duration) (time.Duration, error) {
	if value == "" {
		return fallback, nil
	}
	return time.ParseDuration(value)
}

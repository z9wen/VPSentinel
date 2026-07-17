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
	switch c.GinMode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
	default:
		return fmt.Errorf("unsupported Gin mode %q", c.GinMode)
	}
	return nil
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

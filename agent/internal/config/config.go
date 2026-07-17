package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultInterval       = 15 * time.Second
	defaultRequestTimeout = 10 * time.Second
	defaultQueueCapacity  = 120
)

type Config struct {
	ServerURL         string
	Token             string
	DiskPath          string
	Interval          time.Duration
	RequestTimeout    time.Duration
	QueueCapacity     int
	DryRun            bool
	AllowInsecureHTTP bool
}

func Load(args []string) (Config, error) {
	return load(args, os.Getenv, os.Stderr)
}

func load(args []string, getenv func(string) string, output io.Writer) (Config, error) {
	interval, err := envDuration(getenv, "VPSENTINEL_INTERVAL", defaultInterval)
	if err != nil {
		return Config{}, err
	}
	requestTimeout, err := envDuration(getenv, "VPSENTINEL_REQUEST_TIMEOUT", defaultRequestTimeout)
	if err != nil {
		return Config{}, err
	}
	queueCapacity, err := envInt(getenv, "VPSENTINEL_QUEUE_CAPACITY", defaultQueueCapacity)
	if err != nil {
		return Config{}, err
	}
	dryRun, err := envBool(getenv, "VPSENTINEL_DRY_RUN", false)
	if err != nil {
		return Config{}, err
	}
	allowInsecureHTTP, err := envBool(getenv, "VPSENTINEL_ALLOW_INSECURE_HTTP", false)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	flags := flag.NewFlagSet("vpsentinel-agent", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&config.ServerURL, "server-url", getenv("VPSENTINEL_SERVER_URL"), "metric ingestion URL")
	flags.StringVar(&config.Token, "token", getenv("VPSENTINEL_TOKEN"), "unique node bearer token")
	flags.StringVar(&config.DiskPath, "disk-path", envOrDefault(getenv, "VPSENTINEL_DISK_PATH", "/"), "filesystem path to monitor")
	flags.DurationVar(&config.Interval, "interval", interval, "metric collection interval")
	flags.DurationVar(&config.RequestTimeout, "request-timeout", requestTimeout, "HTTP request timeout")
	flags.IntVar(&config.QueueCapacity, "queue-capacity", queueCapacity, "maximum in-memory samples")
	flags.BoolVar(&config.DryRun, "dry-run", dryRun, "collect once and print JSON without sending")
	flags.BoolVar(&config.AllowInsecureHTTP, "allow-insecure-http", allowInsecureHTTP, "allow sending the agent token over plain HTTP")
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	if flags.NArg() != 0 {
		return Config{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) Validate() error {
	if c.DiskPath == "" {
		return errors.New("disk path must not be empty")
	}
	if c.Interval <= 0 {
		return errors.New("interval must be greater than zero")
	}
	if c.RequestTimeout <= 0 {
		return errors.New("request timeout must be greater than zero")
	}
	if c.QueueCapacity <= 0 {
		return errors.New("queue capacity must be greater than zero")
	}
	if c.DryRun {
		return nil
	}
	if c.ServerURL == "" {
		return errors.New("server URL is required unless dry-run is enabled")
	}
	parsedURL, err := url.ParseRequestURI(c.ServerURL)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return errors.New("server URL must be a valid http or https URL")
	}
	if parsedURL.Scheme == "http" && !c.AllowInsecureHTTP {
		return errors.New("plain HTTP requires --allow-insecure-http")
	}
	if c.Token == "" {
		return errors.New("token is required unless dry-run is enabled")
	}
	return nil
}

func envOrDefault(getenv func(string) string, key, fallback string) string {
	if value := getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(getenv func(string) string, key string, fallback time.Duration) (time.Duration, error) {
	value := getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func envInt(getenv func(string) string, key string, fallback int) (int, error) {
	value := getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func envBool(getenv func(string) string, key string, fallback bool) (bool, error) {
	value := getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

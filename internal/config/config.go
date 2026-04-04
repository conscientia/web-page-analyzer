package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config encapsulates the configuration for the application
type Config struct {
	// Server
	Port string

	// Logging
	LogLevel  string
	LogFormat string

	// Timeouts
	RequestTimeout   time.Duration // for entire request
	FetchTimeout     time.Duration // for how long to wait for target page to respond
	LinkCheckTimeout time.Duration // for total for all link checks
	PerLinkTimeout   time.Duration // for per individual link check

	// Concurrency
	MaxWorkers int // for parallel link checks per request
}

const (
	defaultPort             = "8090"
	defaultRequestTimeout   = 60 * time.Second
	defaultFetchTimeout     = 10 * time.Second
	defaultLinkCheckTimeout = 30 * time.Second
	defaultPerLinkTimeout   = 3 * time.Second
	defaultMaxWorkers       = 15
	defaultLogLevel         = "debug"
	defaultLogFormat        = "text"
)

// Load return the Config from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnvString("PORT", defaultPort),
		LogLevel:         getEnvString("LOG_LEVEL", defaultLogLevel),
		LogFormat:        getEnvString("LOG_FORMAT", defaultLogFormat),
		RequestTimeout:   getEnvDuration("REQUEST_TIMEOUT", defaultRequestTimeout),
		FetchTimeout:     getEnvDuration("FETCH_TIMEOUT", defaultFetchTimeout),
		LinkCheckTimeout: getEnvDuration("LINK_CHECK_TIMEOUT", defaultLinkCheckTimeout),
		PerLinkTimeout:   getEnvDuration("PER_LINK_TIMEOUT", defaultPerLinkTimeout),
		MaxWorkers:       getEnvInt("MAX_WORKERS", defaultMaxWorkers),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// validate enforces the timeouts are proper
func (c *Config) validate() error {
	if c.MaxWorkers < 1 {
		return fmt.Errorf("MAX_WORKERS must be >= 1, got %d", c.MaxWorkers)
	}

	if c.PerLinkTimeout >= c.LinkCheckTimeout {
		return fmt.Errorf(
			"PER_LINK_TIMEOUT (%s) must be < LINK_CHECK_TIMEOUT (%s)",
			c.PerLinkTimeout, c.LinkCheckTimeout,
		)
	}

	if c.FetchTimeout >= c.RequestTimeout {
		return fmt.Errorf(
			"FETCH_TIMEOUT (%s) must be < REQUEST_TIMEOUT (%s)",
			c.FetchTimeout, c.RequestTimeout,
		)
	}

	if c.FetchTimeout+c.LinkCheckTimeout >= c.RequestTimeout {
		return fmt.Errorf(
			"FETCH_TIMEOUT (%s) + LINK_CHECK_TIMEOUT (%s) must be < REQUEST_TIMEOUT (%s)",
			c.FetchTimeout, c.LinkCheckTimeout, c.RequestTimeout,
		)
	}

	return nil
}

// String returns the config.
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Port:%s LogLevel:%s LogFormat:%s RequestTimeout:%s FetchTimeout:%s LinkCheckTimeout:%s PerLinkTimeout:%s MaxWorkers:%d}",
		c.Port,
		c.LogLevel,
		c.LogFormat,
		c.RequestTimeout,
		c.FetchTimeout,
		c.LinkCheckTimeout,
		c.PerLinkTimeout,
		c.MaxWorkers,
	)
}

func getEnvString(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultVal
	}
	return d
}

func getEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

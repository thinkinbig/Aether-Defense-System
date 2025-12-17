package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds configuration for Snowflake ID generation.
type Config struct {
	WorkerID int64 `json:"worker_id" yaml:"worker_id"`
}

// NewConfigFromEnv creates a Snowflake configuration from environment variables.
// It supports the following environment variables:
//   - SNOWFLAKE_WORKER_ID: Direct worker ID specification
//   - HOSTNAME: Kubernetes pod hostname (used to extract StatefulSet ordinal)
//   - POD_NAME: Kubernetes pod name (alternative to HOSTNAME)
//
// If none of these are set, it defaults to worker ID 0.
func NewConfigFromEnv() (*Config, error) {
	// Try direct worker ID specification first
	if workerIDStr := os.Getenv("SNOWFLAKE_WORKER_ID"); workerIDStr != "" {
		workerID, err := strconv.ParseInt(workerIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid SNOWFLAKE_WORKER_ID: %w", err)
		}
		if workerID < 0 || workerID > MaxWorkerID {
			return nil, fmt.Errorf("SNOWFLAKE_WORKER_ID must be between 0 and %d, got %d", MaxWorkerID, workerID)
		}
		return &Config{WorkerID: workerID}, nil
	}

	// Try to extract worker ID from Kubernetes pod name
	if workerID, err := extractWorkerIDFromPodName(); err == nil {
		return &Config{WorkerID: workerID}, nil
	}

	// Default to worker ID 0 with a warning
	_, _ = fmt.Fprintf(os.Stderr, "Warning: No worker ID configuration found, using default worker ID 0. "+
		"Set SNOWFLAKE_WORKER_ID or ensure proper Kubernetes StatefulSet naming.\n")
	return &Config{WorkerID: 0}, nil
}

// extractWorkerIDFromPodName attempts to extract a worker ID from the Kubernetes pod name.
// It looks for StatefulSet naming patterns like "service-name-0", "service-name-1", etc.
func extractWorkerIDFromPodName() (int64, error) {
	// Try HOSTNAME first (usually set by Kubernetes)
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		// Fallback to POD_NAME
		hostname = os.Getenv("POD_NAME")
	}

	if hostname == "" {
		return 0, fmt.Errorf("no hostname or pod name available")
	}

	// Look for StatefulSet pattern: service-name-{ordinal}
	parts := strings.Split(hostname, "-")
	if len(parts) < 2 {
		return 0, fmt.Errorf("hostname does not match StatefulSet pattern: %s", hostname)
	}

	// Try to parse the last part as ordinal
	ordinalStr := parts[len(parts)-1]
	ordinal, err := strconv.ParseInt(ordinalStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ordinal from hostname %s: %w", hostname, err)
	}

	// Ensure ordinal is within valid worker ID range
	if ordinal < 0 || ordinal > MaxWorkerID {
		return 0, fmt.Errorf("ordinal %d from hostname %s is outside valid worker ID range (0-%d)",
			ordinal, hostname, MaxWorkerID)
	}

	return ordinal, nil
}

// NewGenerator creates a new Snowflake generator from the configuration.
func (c *Config) NewGenerator() (*Generator, error) {
	return NewGenerator(c.WorkerID)
}

// MustNewGenerator creates a new Snowflake generator from the configuration and panics on error.
func (c *Config) MustNewGenerator() *Generator {
	gen, err := c.NewGenerator()
	if err != nil {
		panic(fmt.Sprintf("failed to create snowflake generator: %v", err))
	}
	return gen
}

// InitializeDefault initializes the default generator with configuration from environment.
// This should be called once during application startup.
func InitializeDefault() error {
	config, err := NewConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create snowflake config: %w", err)
	}

	generator, err := config.NewGenerator()
	if err != nil {
		return fmt.Errorf("failed to create snowflake generator: %w", err)
	}

	DefaultGenerator = generator
	return nil
}

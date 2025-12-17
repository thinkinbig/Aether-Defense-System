package snowflake

import (
	"os"
	"testing"
)

// Helper function to ignore errors in tests for environment variable operations.
func setEnvIgnoreError(key, value string) {
	// In tests, we ignore environment variable errors as they're not critical
	// and the test environment should be controlled
	_ = os.Setenv(key, value) //nolint:errcheck // Intentionally ignoring errors in tests
}

func unsetEnvIgnoreError(key string) {
	// In tests, we ignore environment variable errors as they're not critical
	// and the test environment should be controlled
	_ = os.Unsetenv(key) //nolint:errcheck // Intentionally ignoring errors in tests
}

func TestNewConfigFromEnv(t *testing.T) {
	// Save original environment
	originalWorkerID := os.Getenv("SNOWFLAKE_WORKER_ID")
	originalHostname := os.Getenv("HOSTNAME")
	originalPodName := os.Getenv("POD_NAME")

	defer func() {
		// Restore original environment
		setEnvIgnoreError("SNOWFLAKE_WORKER_ID", originalWorkerID)
		setEnvIgnoreError("HOSTNAME", originalHostname)
		setEnvIgnoreError("POD_NAME", originalPodName)
	}()

	tests := []struct {
		name             string
		workerID         string
		hostname         string
		podName          string
		expectedWorkerID int64
		expectError      bool
	}{
		{
			name:             "direct worker ID",
			workerID:         "123",
			expectedWorkerID: 123,
			expectError:      false,
		},
		{
			name:        "invalid worker ID",
			workerID:    "1024",
			expectError: true,
		},
		{
			name:        "negative worker ID",
			workerID:    "-1",
			expectError: true,
		},
		{
			name:        "non-numeric worker ID",
			workerID:    "abc",
			expectError: true,
		},
		{
			name:             "hostname StatefulSet pattern",
			hostname:         "trade-rpc-5",
			expectedWorkerID: 5,
			expectError:      false,
		},
		{
			name:             "pod name StatefulSet pattern",
			podName:          "promotion-rpc-42",
			expectedWorkerID: 42,
			expectError:      false,
		},
		{
			name:             "hostname with multiple dashes",
			hostname:         "aether-defense-trade-rpc-7",
			expectedWorkerID: 7,
			expectError:      false,
		},
		{
			name:             "invalid hostname pattern",
			hostname:         "invalid-hostname-abc",
			expectedWorkerID: 0, // defaults to 0
			expectError:      false,
		},
		{
			name:             "no environment variables",
			expectedWorkerID: 0, // defaults to 0
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			unsetEnvIgnoreError("SNOWFLAKE_WORKER_ID")
			unsetEnvIgnoreError("HOSTNAME")
			unsetEnvIgnoreError("POD_NAME")

			// Set test environment
			if tt.workerID != "" {
				setEnvIgnoreError("SNOWFLAKE_WORKER_ID", tt.workerID)
			}
			if tt.hostname != "" {
				setEnvIgnoreError("HOSTNAME", tt.hostname)
			}
			if tt.podName != "" {
				setEnvIgnoreError("POD_NAME", tt.podName)
			}

			config, err := NewConfigFromEnv()

			if tt.expectError {
				if err == nil {
					t.Errorf("NewConfigFromEnv() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewConfigFromEnv() unexpected error: %v", err)
				return
			}

			if config.WorkerID != tt.expectedWorkerID {
				t.Errorf("NewConfigFromEnv() worker ID = %d, want %d", config.WorkerID, tt.expectedWorkerID)
			}
		})
	}
}

func TestConfig_NewGenerator(t *testing.T) {
	config := &Config{WorkerID: 42}

	gen, err := config.NewGenerator()
	if err != nil {
		t.Errorf("Config.NewGenerator() error = %v", err)
		return
	}

	if gen.GetWorkerID() != 42 {
		t.Errorf("Config.NewGenerator() worker ID = %d, want 42", gen.GetWorkerID())
	}
}

func TestConfig_MustNewGenerator(t *testing.T) {
	config := &Config{WorkerID: 42}

	// Should not panic with valid config
	gen := config.MustNewGenerator()
	if gen.GetWorkerID() != 42 {
		t.Errorf("Config.MustNewGenerator() worker ID = %d, want 42", gen.GetWorkerID())
	}
}

func TestConfig_MustNewGenerator_Panic(t *testing.T) {
	config := &Config{WorkerID: -1} // Invalid worker ID

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Config.MustNewGenerator() expected panic with invalid worker ID")
		}
	}()

	config.MustNewGenerator()
}

func TestInitializeDefault(t *testing.T) {
	// Save original environment and default generator
	originalWorkerID := os.Getenv("SNOWFLAKE_WORKER_ID")
	originalDefault := DefaultGenerator

	defer func() {
		// Restore original environment and default generator
		setEnvIgnoreError("SNOWFLAKE_WORKER_ID", originalWorkerID)
		DefaultGenerator = originalDefault
	}()

	// Set test environment
	setEnvIgnoreError("SNOWFLAKE_WORKER_ID", "99")

	err := InitializeDefault()
	if err != nil {
		t.Errorf("InitializeDefault() error = %v", err)
		return
	}

	if DefaultGenerator.GetWorkerID() != 99 {
		t.Errorf("InitializeDefault() default generator worker ID = %d, want 99", DefaultGenerator.GetWorkerID())
	}

	// Test that package functions work with new default
	id, err := Next()
	if err != nil {
		t.Errorf("Next() error after InitializeDefault(): %v", err)
	}

	_, workerID, _ := ParseID(id)
	if workerID != 99 {
		t.Errorf("Generated ID worker ID = %d, want 99", workerID)
	}
}

func TestExtractWorkerIDFromPodName(t *testing.T) {
	// Save original environment
	originalHostname := os.Getenv("HOSTNAME")
	originalPodName := os.Getenv("POD_NAME")

	defer func() {
		// Restore original environment
		setEnvIgnoreError("HOSTNAME", originalHostname)
		setEnvIgnoreError("POD_NAME", originalPodName)
	}()

	tests := []struct {
		name        string
		hostname    string
		podName     string
		expectedID  int64
		expectError bool
	}{
		{
			name:        "valid hostname",
			hostname:    "service-0",
			expectedID:  0,
			expectError: false,
		},
		{
			name:        "valid pod name",
			podName:     "service-123",
			expectedID:  123,
			expectError: false,
		},
		{
			name:        "hostname takes priority",
			hostname:    "service-1",
			podName:     "service-2",
			expectedID:  1,
			expectError: false,
		},
		{
			name:        "no environment",
			expectError: true,
		},
		{
			name:        "invalid pattern",
			hostname:    "service",
			expectError: true,
		},
		{
			name:        "non-numeric ordinal",
			hostname:    "service-abc",
			expectError: true,
		},
		{
			name:        "ordinal too large",
			hostname:    "service-2000",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			unsetEnvIgnoreError("HOSTNAME")
			unsetEnvIgnoreError("POD_NAME")

			// Set test environment
			if tt.hostname != "" {
				setEnvIgnoreError("HOSTNAME", tt.hostname)
			}
			if tt.podName != "" {
				setEnvIgnoreError("POD_NAME", tt.podName)
			}

			workerID, err := extractWorkerIDFromPodName()

			if tt.expectError {
				if err == nil {
					t.Errorf("extractWorkerIDFromPodName() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("extractWorkerIDFromPodName() unexpected error: %v", err)
				return
			}

			if workerID != tt.expectedID {
				t.Errorf("extractWorkerIDFromPodName() = %d, want %d", workerID, tt.expectedID)
			}
		})
	}
}

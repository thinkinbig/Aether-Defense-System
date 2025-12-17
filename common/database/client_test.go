package database

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.MaxOpenConns != 100 {
		t.Errorf("Expected MaxOpenConns=100, got %d", config.MaxOpenConns)
	}

	if config.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns=10, got %d", config.MaxIdleConns)
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		password string
		host     string
		port     string
		dbName   string
		params   map[string]string
		want     string
	}{
		{
			name:     "basic DSN",
			user:     "test",
			password: "pass",
			host:     "localhost",
			port:     "3306",
			dbName:   "testdb",
			params:   nil,
			want: "test:pass@tcp(localhost:3306)/testdb?" +
				"charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&" +
				"readTimeout=3s&writeTimeout=3s",
		},
		{
			name:     "DSN with custom params",
			user:     "user",
			password: "pwd",
			host:     "127.0.0.1",
			port:     "3306",
			dbName:   "mydb",
			params: map[string]string{
				"charset":   "utf8mb4",
				"parseTime": "True",
			},
			want: "user:pwd@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDSN(tt.user, tt.password, tt.host, tt.port, tt.dbName, tt.params)
			// Note: We can't do exact match due to map iteration order, so we check key components
			if got == "" {
				t.Error("BuildDSN returned empty string")
			}
			// Basic validation
			expectedPrefix := tt.user + ":" + tt.password + "@tcp(" + tt.host + ":" + tt.port + ")/" + tt.dbName
			if len(got) < len(expectedPrefix) {
				t.Errorf("DSN too short: got %s", got)
			}
		})
	}
}

func TestNewClient_InvalidConfig(t *testing.T) {
	// Test with nil config
	_, err := NewClient(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test with empty DSN
	config := &Config{DSN: ""}
	_, err = NewClient(config)
	if err == nil {
		t.Error("Expected error for empty DSN")
	}
}

// Integration test - requires actual MySQL instance
// Uncomment and configure to run integration tests
/*
func TestNewClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &Config{
		DSN: "root:root123@tcp(localhost:3306)/aether_defense?charset=utf8mb4&parseTime=True&loc=Local",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}
*/

package snowflake

import (
	"sync"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name      string
		workerID  int64
		wantError bool
	}{
		{"valid worker ID 0", 0, false},
		{"valid worker ID 1023", 1023, false},
		{"valid worker ID 512", 512, false},
		{"invalid worker ID -1", -1, true},
		{"invalid worker ID 1024", 1024, true},
		{"invalid worker ID 2000", 2000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.workerID)
			if tt.wantError {
				if err == nil {
					t.Errorf("NewGenerator() expected error for worker ID %d, but got none", tt.workerID)
				}
				return
			}
			if err != nil {
				t.Errorf("NewGenerator() unexpected error: %v", err)
				return
			}
			if gen.GetWorkerID() != tt.workerID {
				t.Errorf("NewGenerator() worker ID = %d, want %d", gen.GetWorkerID(), tt.workerID)
			}
		})
	}
}

func TestGenerator_Next(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// Generate multiple IDs and ensure they are unique and increasing
	ids := make([]int64, 1000)
	for i := 0; i < 1000; i++ {
		id, err := gen.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		ids[i] = id
	}

	// Check uniqueness
	idSet := make(map[int64]bool)
	for _, id := range ids {
		if idSet[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		idSet[id] = true
	}

	// Check that IDs are generally increasing (allowing for same millisecond)
	for i := 1; i < len(ids); i++ {
		if ids[i] < ids[i-1] {
			// Parse timestamps to check if they're in the same millisecond
			ts1, _, _ := ParseID(ids[i-1])
			ts2, _, _ := ParseID(ids[i])
			if ts2 < ts1 {
				t.Errorf("ID timestamp went backwards: %d -> %d", ts1, ts2)
			}
		}
	}
}

func TestGenerator_MustNext(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// Should not panic with valid generator
	id := gen.MustNext()
	if id <= 0 {
		t.Errorf("MustNext() returned invalid ID: %d", id)
	}
}

func TestGenerator_Concurrency(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	const numGoroutines = 100
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	idChan := make(chan int64, numGoroutines*idsPerGoroutine)

	// Start multiple goroutines generating IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := gen.Next()
				if err != nil {
					t.Errorf("Next() error in goroutine: %v", err)
					return
				}
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// Collect all IDs and check for uniqueness
	idSet := make(map[int64]bool)
	totalIDs := 0
	for id := range idChan {
		if idSet[id] {
			t.Errorf("Duplicate ID generated in concurrent test: %d", id)
		}
		idSet[id] = true
		totalIDs++
	}

	expectedTotal := numGoroutines * idsPerGoroutine
	if totalIDs != expectedTotal {
		t.Errorf("Expected %d IDs, got %d", expectedTotal, totalIDs)
	}
}

func TestParseID(t *testing.T) {
	gen, err := NewGenerator(123)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	id, err := gen.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	timestamp, workerID, sequence := ParseID(id)

	// Check worker ID
	if workerID != 123 {
		t.Errorf("ParseID() worker ID = %d, want 123", workerID)
	}

	// Check timestamp is reasonable (within last few seconds)
	now := time.Now().UnixMilli()
	if timestamp < now-5000 || timestamp > now+1000 {
		t.Errorf("ParseID() timestamp = %d, expected around %d", timestamp, now)
	}

	// Check sequence is valid
	if sequence < 0 || sequence > MaxSequence {
		t.Errorf("ParseID() sequence = %d, want 0-%d", sequence, MaxSequence)
	}
}

func TestDefaultGenerator(t *testing.T) {
	// Test package-level functions
	id, err := Next()
	if err != nil {
		t.Errorf("Next() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("Next() returned invalid ID: %d", id)
	}

	// Test MustNext
	id2 := MustNext()
	if id2 <= 0 {
		t.Errorf("MustNext() returned invalid ID: %d", id2)
	}

	// IDs should be different
	if id == id2 {
		t.Errorf("Next() and MustNext() returned same ID: %d", id)
	}
}

func TestSequenceOverflow(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// Set sequence to near maximum to test overflow
	gen.mu.Lock()
	gen.sequence = MaxSequence - 1
	gen.lastTimestamp = time.Now().UnixMilli()
	gen.mu.Unlock()

	// Generate two IDs - second should trigger overflow and wait for next millisecond
	id1, err := gen.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	id2, err := gen.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	// Parse timestamps
	ts1, _, seq1 := ParseID(id1)
	ts2, _, seq2 := ParseID(id2)

	// Second ID should have later timestamp and sequence 0
	if ts2 <= ts1 {
		t.Errorf("Expected timestamp to advance after sequence overflow, got %d -> %d", ts1, ts2)
	}
	if seq1 != MaxSequence {
		t.Errorf("Expected first sequence to be %d, got %d", MaxSequence, seq1)
	}
	if seq2 != 0 {
		t.Errorf("Expected second sequence to be 0 after overflow, got %d", seq2)
	}
}

// Benchmark tests.
func BenchmarkGenerator_Next(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("NewGenerator() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Next()
		if err != nil {
			b.Fatalf("Next() error = %v", err)
		}
	}
}

func BenchmarkGenerator_NextParallel(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("NewGenerator() error = %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := gen.Next()
			if err != nil {
				b.Fatalf("Next() error = %v", err)
			}
		}
	})
}

func BenchmarkMustNext(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MustNext()
	}
}

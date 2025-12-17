// Package snowflake implements distributed unique ID generation using the Snowflake algorithm.
//
// The Snowflake algorithm generates 64-bit unique IDs with the following structure:
// - 1 bit: Sign bit (always 0)
// - 41 bits: Timestamp (milliseconds since epoch)
// - 10 bits: Worker ID (supports up to 1024 nodes)
// - 12 bits: Sequence number (up to 4096 IDs per millisecond per worker)
//
// This implementation is optimized for high-concurrency scenarios and ensures
// trend-increasing IDs for optimal B+ tree insertion performance in MySQL.
package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	// Epoch is the custom epoch for this system (2023-01-01 00:00:00 UTC)
	// This extends the usable lifetime of the algorithm by ~69 years from this date.
	Epoch int64 = 1672531200000 // 2023-01-01 00:00:00 UTC in milliseconds

	// WorkerIDBits defines the number of bits allocated for worker ID (10 bits for 0-1023).
	WorkerIDBits = 10
	// SequenceBits defines the number of bits allocated for sequence (12 bits for 0-4095).
	SequenceBits = 12

	// MaxWorkerID is the maximum worker ID value (1023).
	MaxWorkerID = (1 << WorkerIDBits) - 1
	// MaxSequence is the maximum sequence value (4095).
	MaxSequence = (1 << SequenceBits) - 1

	// WorkerIDShift is the bit shift for worker ID (12).
	WorkerIDShift = SequenceBits
	// TimestampShift is the bit shift for timestamp (22).
	TimestampShift = SequenceBits + WorkerIDBits
)

// Generator represents a Snowflake ID generator instance.
type Generator struct {
	mu            sync.Mutex
	workerID      int64
	sequence      int64
	lastTimestamp int64
}

// NewGenerator creates a new Snowflake ID generator with the specified worker ID.
// The worker ID must be between 0 and 1023 (inclusive).
func NewGenerator(workerID int64) (*Generator, error) {
	if workerID < 0 || workerID > MaxWorkerID {
		return nil, fmt.Errorf("worker ID must be between 0 and %d, got %d", MaxWorkerID, workerID)
	}

	return &Generator{
		workerID: workerID,
	}, nil
}

// Next generates the next unique ID.
// This method is thread-safe and can be called concurrently.
func (g *Generator) Next() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := g.getCurrentTimestamp()

	// Handle clock going backwards
	if timestamp < g.lastTimestamp {
		return 0, fmt.Errorf("clock moved backwards, refusing to generate ID for %d milliseconds",
			g.lastTimestamp-timestamp)
	}

	// Same millisecond as last ID generation
	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & MaxSequence

		// Sequence overflow - wait for next millisecond
		if g.sequence == 0 {
			timestamp = g.waitNextMillis(g.lastTimestamp)
		}
	} else {
		// New millisecond - reset sequence
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	// Construct the ID
	id := ((timestamp - Epoch) << TimestampShift) |
		(g.workerID << WorkerIDShift) |
		g.sequence

	return id, nil
}

// MustNext generates the next unique ID and panics on error.
// Use this method only when you're certain the generator is properly configured.
func (g *Generator) MustNext() int64 {
	id, err := g.Next()
	if err != nil {
		panic(fmt.Sprintf("failed to generate snowflake ID: %v", err))
	}
	return id
}

// getCurrentTimestamp returns the current timestamp in milliseconds.
func (g *Generator) getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// waitNextMillis waits until the next millisecond.
func (g *Generator) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := g.getCurrentTimestamp()
	for timestamp <= lastTimestamp {
		timestamp = g.getCurrentTimestamp()
	}
	return timestamp
}

// ParseID parses a Snowflake ID and returns its components.
func ParseID(id int64) (timestamp, workerID, sequence int64) {
	timestamp = (id >> TimestampShift) + Epoch
	workerID = (id >> WorkerIDShift) & MaxWorkerID
	sequence = id & MaxSequence
	return
}

// GetWorkerID returns the worker ID of this generator.
func (g *Generator) GetWorkerID() int64 {
	return g.workerID
}

// DefaultGenerator is a package-level generator instance for convenience.
// It uses worker ID 0 and should be replaced with a properly configured
// generator in production environments.
var DefaultGenerator *Generator

func init() {
	var err error
	DefaultGenerator, err = NewGenerator(0)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize default snowflake generator: %v", err))
	}
}

// Next generates the next unique ID using the default generator.
func Next() (int64, error) {
	return DefaultGenerator.Next()
}

// MustNext generates the next unique ID using the default generator and panics on error.
func MustNext() int64 {
	return DefaultGenerator.MustNext()
}

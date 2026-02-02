package collector

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	bpfsv1 "github.com/Tjaarda1/bpfstats/api/v1"
	"github.com/cilium/ebpf"
)

type LatencyCollector struct {
	id       uint32
	s        *Stats
	interval time.Duration

	// Lifecycle management
	mu      sync.RWMutex
	running bool
	done    chan struct{}
	errCh   chan error

	// Measurement metadata
	started time.Time
	warmup  *time.Duration
}

// NewLatencyCollector creates a new latency collector
func NewLatencyCollector(id uint32, interval time.Duration, warmup *time.Duration) *LatencyCollector {
	return &LatencyCollector{
		id:       id,
		s:        &Stats{},
		interval: interval,
		warmup:   warmup,
		done:     make(chan struct{}),
		errCh:    make(chan error, 1), // buffered to prevent goroutine leak
	}
}

// Start begins collecting latency statistics
func (latC *LatencyCollector) Start(ctx context.Context) error {
	latC.mu.Lock()
	if latC.running {
		latC.mu.Unlock()
		return fmt.Errorf("collector already running")
	}
	latC.running = true
	latC.started = time.Now()
	latC.mu.Unlock()

	ticker := time.NewTicker(latC.interval)
	defer ticker.Stop()

	warmupEnd := latC.started
	if latC.warmup != nil {
		warmupEnd = latC.started.Add(*latC.warmup)
	}

	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown pattern from net/http.Server
			latC.mu.Lock()
			latC.running = false
			latC.mu.Unlock()
			close(latC.done)
			return ctx.Err()

		case <-latC.done:
			// Explicit Stop() called
			return nil

		case <-ticker.C:
			// Collect sample
			prog, err := ebpf.NewProgramFromID(ebpf.ProgramID(latC.id))
			if err != nil {
				// Non-fatal error handling inspired by Prometheus
				select {
				case latC.errCh <- fmt.Errorf("NewProgramFromID: %w", err):
				default:
					// Don't block if error channel is full
				}
				continue
			}

			stats, err := prog.Stats()
			prog.Close() // Close immediately after use
			if err != nil {
				select {
				case latC.errCh <- fmt.Errorf("Stats: %w", err):
				default:
				}
				continue
			}

			// Skip samples during warmup period
			if latC.warmup != nil && time.Now().Before(warmupEnd) {
				continue
			}

			// Record latency sample (runtime per invocation in nanoseconds)
			if stats.RunCount > 0 {
				avgLatencyNs := float64(stats.Runtime) / float64(stats.RunCount)
				latC.s.Add(avgLatencyNs)
			}
		}
	}
}

// Stop gracefully stops the collector
func (latC *LatencyCollector) Stop() error {
	latC.mu.Lock()
	defer latC.mu.Unlock()

	if !latC.running {
		return fmt.Errorf("collector not running")
	}

	latC.running = false
	close(latC.done)

	return nil
}

// Snapshot captures current statistics without stopping collection
func (latC *LatencyCollector) Snapshot() (bpfsv1.Parameter, error) {
	latC.mu.RLock()
	defer latC.mu.RUnlock()

	// Thread-safe read from Stats
	count := latC.s.Count()
	if count == 0 {
		return nil, fmt.Errorf("no samples collected yet")
	}

	mean := latC.s.Mean()
	variance := latC.s.Variance()
	stddev := math.Sqrt(variance)

	min := uint64(latC.s.Min())
	max := uint64(latC.s.Max())

	// Coefficient of variation
	cv := stddev / mean

	now := time.Now()
	duration := now.Sub(latC.started)

	// Adjust duration if warmup was used
	if latC.warmup != nil {
		duration -= *latC.warmup
		if duration < 0 {
			duration = 0
		}
	}

	rate := float64(count) / duration.Seconds()

	latency := bpfsv1.Latency{
		ID:       latC.id,
		Duration: duration,
		Warmup:   latC.warmup,
		Started:  &latC.started,
		Ended:    &now,

		Samples: count,
		Rate:    &rate,

		Mean:   uint64(mean),
		StdDev: uint64(stddev),
		CV:     &cv,
		Min:    &min,
		Max:    &max,

		// Note: Percentiles would require storing all samples or using a sketch
		// For now, leaving it nil - see below for extension
		Percentiles: nil,
	}

	return latency, nil
}

// Err returns the most recent non-fatal error (if any)
// Pattern from: Go's sql.DB
func (latC *LatencyCollector) Err() error {
	select {
	case err := <-latC.errCh:
		return err
	default:
		return nil
	}
}

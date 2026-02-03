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

type CpuCollector struct {
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

// NewCpuCollector creates a new cpu collector
func NewCPUCollector(id uint32, interval time.Duration, warmup *time.Duration) *CpuCollector {
	return &CpuCollector{
		id:       id,
		s:        &Stats{},
		interval: interval,
		warmup:   warmup,
		done:     make(chan struct{}),
		errCh:    make(chan error, 1), // buffered to prevent goroutine leak
	}
}

// Start begins collecting cpu statistics
func (cpuC *CpuCollector) Start(ctx context.Context) error {
	cpuC.mu.Lock()
	if cpuC.running {
		cpuC.mu.Unlock()
		return fmt.Errorf("collector already running")
	}
	cpuC.running = true
	cpuC.started = time.Now()
	cpuC.mu.Unlock()

	ticker := time.NewTicker(cpuC.interval)
	defer ticker.Stop()

	warmupEnd := cpuC.started
	if cpuC.warmup != nil {
		warmupEnd = cpuC.started.Add(*cpuC.warmup)
	}

	lastRuntime := time.Nanosecond * 0
	var lastTime *time.Time
	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown pattern from net/http.Server
			cpuC.mu.Lock()
			cpuC.running = false
			cpuC.mu.Unlock()
			// close(cpuC.done)
			return ctx.Err()

		case <-cpuC.done:
			// Explicit Stop() called
			return nil

		case <-ticker.C:
			// Collect sample
			prog, err := ebpf.NewProgramFromID(ebpf.ProgramID(cpuC.id))
			if err != nil {
				// Non-fatal error handling inspired by Prometheus
				select {
				case cpuC.errCh <- fmt.Errorf("NewProgramFromID: %w", err):
				default:
					// Don't block if error channel is full
				}
				continue
			}

			stats, err := prog.Stats()
			prog.Close() // Close immediately after use
			if err != nil {
				select {
				case cpuC.errCh <- fmt.Errorf("Stats: %w", err):
				default:
				}
				continue
			}

			// Skip samples during warmup period
			if cpuC.warmup != nil && time.Now().Before(warmupEnd) {
				continue
			}

			if lastTime == nil {
				n := time.Now()
				lastTime = &n
			}
			// Record cpu sample if there is a new entry
			if stats.RunCount > 0 && stats.Runtime-lastRuntime != 0 {
				now := time.Now()
				dWall := now.Sub(*lastTime).Nanoseconds() // ns

				avgCpuPerc := (float64((stats.Runtime)-lastRuntime) / float64(dWall)) * 100
				cpuC.s.Add(avgCpuPerc)
				lastTime = &now
				lastRuntime = stats.Runtime
			}
		}
	}
}

// Stop gracefully stops the collector
func (cpuC *CpuCollector) Stop() error {
	cpuC.mu.Lock()
	defer cpuC.mu.Unlock()

	if !cpuC.running {
		return nil
	}

	cpuC.running = false
	close(cpuC.done)

	return nil
}

// Snapshot captures current statistics without stopping collection
func (cpuC *CpuCollector) Snapshot() (bpfsv1.Parameter, error) {
	cpuC.mu.RLock()
	defer cpuC.mu.RUnlock()

	// Thread-safe read from Stats
	count := cpuC.s.Count()
	if count == 0 {
		return nil, fmt.Errorf("no samples collected yet")
	}

	mean := cpuC.s.Mean()
	variance := cpuC.s.Variance()
	stddev := math.Sqrt(variance)

	min := uint64(cpuC.s.Min())
	max := uint64(cpuC.s.Max())

	// Coefficient of variation
	cv := stddev / mean

	now := time.Now()
	duration := now.Sub(cpuC.started)

	// Adjust duration if warmup was used
	if cpuC.warmup != nil {
		duration -= *cpuC.warmup
		if duration < 0 {
			duration = 0
		}
	}

	rate := float64(count) / duration.Seconds()

	cpu := bpfsv1.Cpu{
		ID:       cpuC.id,
		Duration: duration,
		Warmup:   cpuC.warmup,
		Started:  &cpuC.started,
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
	}

	return cpu, nil
}

// Err returns the most recent non-fatal error (if any)
// Pattern from:t Go's sql.DB
func (cpuC *CpuCollector) Err() error {
	select {
	case err := <-cpuC.errCh:
		return err
	default:
		return nil
	}
}

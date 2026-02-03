package v1

import "time"

type Parameter interface {
	Kind() string
}

func (Latency) Kind() string { return "latency" }
func (Cpu) Kind() string     { return "cpu" }

// Latency is a Parameter payload containing distribution-aware latency statistics.
// Units: all duration-like fields are nanoseconds unless otherwise stated.
type Latency struct {
	// Identity / target
	ID uint32 `json:"id"` // e.g. kernel bpf program id (or bpfstat target id)

	// Measurement window
	Duration time.Duration  `json:"duration"`          // e.g. 60s
	Warmup   *time.Duration `json:"warmup,omitempty"`  // if you support warmup/discard
	Started  *time.Time     `json:"started,omitempty"` // optional metadata
	Ended    *time.Time     `json:"ended,omitempty"`   // optional metadata

	// Volume / integrity
	Samples uint64   `json:"samples"`                // n
	Dropped *uint64  `json:"dropped,omitempty"`      // lost events / ringbuf drops etc.
	Rate    *float64 `json:"rate_per_sec,omitempty"` // samples/sec, if computed

	// Summary stats (nanoseconds)
	Mean   uint64   `json:"mean_ns"`      // avg
	StdDev uint64   `json:"stddev_ns"`    // standard deviation
	CV     *float64 `json:"cv,omitempty"` // coefficient of variation (stddev/mean)
	Min    *uint64  `json:"min_ns,omitempty"`
	Max    *uint64  `json:"max_ns,omitempty"`

	// Percentiles in nanoseconds: keys like "p50", "p90", "p99", "p99_9"
	Percentiles *map[string]uint64 `json:"percentiles_ns,omitempty"`

	// Measurement semantics / reproducibility
	Clock     *string `json:"clock,omitempty"`     // e.g. "ktime_ns", "cycles"
	Histogram *string `json:"histogram,omitempty"` // e.g. "log2", "ddsketch", etc.

}

type Cpu struct {
	// Identity / target
	ID uint32 `json:"id"` // e.g. kernel bpf program id (or bpfstat target id)

	// Measurement window
	Duration time.Duration  `json:"duration"`          // e.g. 60s
	Warmup   *time.Duration `json:"warmup,omitempty"`  // if you support warmup/discard
	Started  *time.Time     `json:"started,omitempty"` // optional metadata
	Ended    *time.Time     `json:"ended,omitempty"`   // optional metadata

	// Volume / integrity
	Samples uint64   `json:"samples"`                // n
	Dropped *uint64  `json:"dropped,omitempty"`      // lost events / ringbuf drops etc.
	Rate    *float64 `json:"rate_per_sec,omitempty"` // samples/sec, if computed

	// Summary stats (nanoseconds)
	Mean   uint64   `json:"mean_perc"`    // avg
	StdDev uint64   `json:"stddev_perc"`  // standard deviation
	CV     *float64 `json:"cv,omitempty"` // coefficient of variation (stddev/mean)
	Min    *uint64  `json:"min_perc,omitempty"`
	Max    *uint64  `json:"max_perc,omitempty"`

	// Measurement semantics / reproducibility
	Clock     *string `json:"clock,omitempty"`     // e.g. "ktime_ns", "cycles"
	Histogram *string `json:"histogram,omitempty"` // e.g. "log2", "ddsketch", etc.

}

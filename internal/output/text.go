package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	bpfsv1 "github.com/Tjaarda1/bpfstats/api/v1"
)

type TextOutput struct{}

func (t *TextOutput) OutputParam(par bpfsv1.Parameter, w io.Writer) error {
	switch par.Kind() {
	case "latency":
		return t.outputLatency(par.(bpfsv1.Latency), w)
	case "cpu":
		return t.outputCpu(par.(bpfsv1.Cpu), w)
	default:
		return fmt.Errorf("unsupported parameter kind: %s", par.Kind())
	}
}

func (t *TextOutput) outputLatency(lat bpfsv1.Latency, w io.Writer) error {
	var sb strings.Builder

	// Header
	sb.WriteString("=== Latency Statistics ===\n\n")

	// Identity
	sb.WriteString(fmt.Sprintf("ID: %d\n", lat.ID))

	// Measurement window
	sb.WriteString(fmt.Sprintf("Duration: %s\n", lat.Duration))
	if lat.Warmup != nil {
		sb.WriteString(fmt.Sprintf("Warmup: %s\n", *lat.Warmup))
	}
	if lat.Started != nil {
		sb.WriteString(fmt.Sprintf("Started: %s\n", lat.Started.Format(time.RFC3339)))
	}
	if lat.Ended != nil {
		sb.WriteString(fmt.Sprintf("Ended: %s\n", lat.Ended.Format(time.RFC3339)))
	}
	sb.WriteString("\n")

	// Volume / integrity
	sb.WriteString(fmt.Sprintf("Samples: %d\n", lat.Samples))
	if lat.Dropped != nil {
		sb.WriteString(fmt.Sprintf("Dropped: %d\n", *lat.Dropped))
	}
	if lat.Rate != nil {
		sb.WriteString(fmt.Sprintf("Rate: %.2f samples/sec\n", *lat.Rate))
	}
	sb.WriteString("\n")

	// Summary stats
	sb.WriteString("--- Summary Statistics ---\n")
	sb.WriteString(fmt.Sprintf("Mean: %s\n", formatNanos(lat.Mean)))
	sb.WriteString(fmt.Sprintf("StdDev: %s\n", formatNanos(lat.StdDev)))
	if lat.CV != nil {
		sb.WriteString(fmt.Sprintf("CV: %.4f\n", *lat.CV))
	}
	if lat.Min != nil {
		sb.WriteString(fmt.Sprintf("Min: %s\n", formatNanos(*lat.Min)))
	}
	if lat.Max != nil {
		sb.WriteString(fmt.Sprintf("Max: %s\n", formatNanos(*lat.Max)))
	}
	sb.WriteString("\n")

	// Percentiles
	if lat.Percentiles != nil && len(*lat.Percentiles) > 0 {
		sb.WriteString("--- Percentiles ---\n")
		// Common ordering for percentiles
		order := []string{"p50", "p90", "p95", "p99", "p99_9", "p99_99"}
		for _, key := range order {
			if val, ok := (*lat.Percentiles)[key]; ok {
				sb.WriteString(fmt.Sprintf("%s: %s\n", key, formatNanos(val)))
			}
		}
		// Print any remaining percentiles not in standard order
		for key, val := range *lat.Percentiles {
			if !contains(order, key) {
				sb.WriteString(fmt.Sprintf("%s: %s\n", key, formatNanos(val)))
			}
		}
		sb.WriteString("\n")
	}

	// Metadata
	if lat.Clock != nil || lat.Histogram != nil {
		sb.WriteString("--- Measurement Info ---\n")
		if lat.Clock != nil {
			sb.WriteString(fmt.Sprintf("Clock: %s\n", *lat.Clock))
		}
		if lat.Histogram != nil {
			sb.WriteString(fmt.Sprintf("Histogram: %s\n", *lat.Histogram))
		}
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

// formatNanos converts nanoseconds to a human-readable duration string
func formatNanos(ns uint64) string {
	d := time.Duration(ns)
	// Format nicely based on magnitude
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", ns)
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fµs", float64(ns)/1000.0)
	case d < time.Second:
		return fmt.Sprintf("%.2fms", float64(ns)/1000000.0)
	default:
		return fmt.Sprintf("%.3fs", float64(ns)/1000000000.0)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (t *TextOutput) outputCpu(cpu bpfsv1.Cpu, w io.Writer) error {
	var sb strings.Builder

	// Header
	sb.WriteString("=== CPU Statistics ===\n\n")

	// Identity
	sb.WriteString(fmt.Sprintf("ID: %d\n", cpu.ID))

	// Measurement window
	sb.WriteString(fmt.Sprintf("Duration: %s\n", cpu.Duration))
	if cpu.Warmup != nil {
		sb.WriteString(fmt.Sprintf("Warmup: %s\n", *cpu.Warmup))
	}
	if cpu.Started != nil {
		sb.WriteString(fmt.Sprintf("Started: %s\n", cpu.Started.Format(time.RFC3339)))
	}
	if cpu.Ended != nil {
		sb.WriteString(fmt.Sprintf("Ended: %s\n", cpu.Ended.Format(time.RFC3339)))
	}
	sb.WriteString("\n")

	// Volume / integrity
	sb.WriteString(fmt.Sprintf("Samples: %d\n", cpu.Samples))
	if cpu.Dropped != nil {
		sb.WriteString(fmt.Sprintf("Dropped: %d\n", *cpu.Dropped))
	}
	if cpu.Rate != nil {
		sb.WriteString(fmt.Sprintf("Rate: %.2f samples/sec\n", *cpu.Rate))
	}
	sb.WriteString("\n")

	// Summary stats (stored as ratios: 0..1)
	sb.WriteString("--- Summary Statistics ---\n")
	sb.WriteString(fmt.Sprintf("Mean: %.2f%%\n", 100.0*cpu.Mean))
	sb.WriteString(fmt.Sprintf("StdDev: %.2f%%\n", 100.0*cpu.StdDev))

	if cpu.CV != nil {
		sb.WriteString(fmt.Sprintf("CV: %.4f\n", *cpu.CV))
	}

	if cpu.Min != nil {
		sb.WriteString(fmt.Sprintf("Min: %.2f%%\n", 100.0*(*cpu.Min)))
	}

	if cpu.Max != nil {
		sb.WriteString(fmt.Sprintf("Max: %.2f%%\n", 100.0*(*cpu.Max)))
	}

	sb.WriteString("\n")

	// Metadata
	if cpu.Clock != nil || cpu.Histogram != nil {
		sb.WriteString("--- Measurement Info ---\n")
		if cpu.Clock != nil {
			sb.WriteString(fmt.Sprintf("Clock: %s\n", *cpu.Clock))
		}
		if cpu.Histogram != nil {
			sb.WriteString(fmt.Sprintf("Histogram: %s\n", *cpu.Histogram))
		}
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	bpfsv1 "github.com/Tjaarda1/bpfstats/api/v1"
	"github.com/Tjaarda1/bpfstats/internal/collector"
	"github.com/Tjaarda1/bpfstats/internal/output"
	"github.com/spf13/cobra"
)

// latencyCmd represents the latency command
func NewCmdLatency(parent string) *cobra.Command {
	flags := NewLatencyFlags()
	cmd := &cobra.Command{
		Use:                   "latency",
		DisableFlagsInUseLine: true,
		Short:                 latencyShort,
		Long:                  latencyLong,
		Example:               latencyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToOptions(parent, args)
			if err != nil {
				return err
			}
			return o.Run()
		},
	}

	flags.AddFlags(cmd)

	return cmd
}
func init() {

	rootCmd.AddCommand(NewCmdLatency(rootCmd.Name()))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// latencyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// latencyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var (
	latencyLong = `
		Measure and report latency statistics for a specific eBPF program.

		This command runs a time-bounded measurement for the selected program ID and prints
		distribution-aware latency statistics, such as mean, standard deviation, and tail
		percentiles (p50/p90/p99/p99.9). The default output is human-readable text; use --json
		for machine-readable output.

		The reported "latency" is the per-invocation execution duration of the eBPF program
		(i.e., time spent executing BPF instructions and helper calls for each trigger), not
		end-to-end application latency. The report includes sample counts and measurement
		integrity metadata (e.g., dropped samples) when available.`

	latencyExample = ` 
		# Measure latency statistics for eBPF program id 42 for 60 seconds
		bpfstat latency --id 42 --duration 60s

		# Same measurement, output as JSON
		bpfstat latency --id 42 --duration 60s --json

		# Write JSON output to a file
		bpfstat latency --id 42 --duration 60s --json -o latency_42.json

		# Use a shorter interval for quick iteration
		bpfstat latency --id 42 --duration 10s

		# Measure with custom percentiles (if supported by your flags)
		bpfstat latency --id 42 --duration 60s --percentiles 50,90,99,99.9`
	latencyShort = "Measure and report latency statistics for a specific eBPF program."
)

// Flags will be converted to options, which are taken when measuring and outputting the stats data
type LatencyFlags struct {

	// Target selection
	ID uint32

	// Measurement window
	Duration time.Duration
	Warmup   time.Duration

	// Output selection
	JSON   bool
	Pretty bool
	Output string // -o / --output file path (empty => stdout)

	// Stats config
	Percentiles []string // e.g. ["50","90","99","99.9"] or ["p50","p99"]

}

// NewLatencyFlags returns a default LatencyFlags
func NewLatencyFlags() *LatencyFlags {
	return &LatencyFlags{}
}

// AddFlags registers flags for a cli
func (flags *LatencyFlags) AddFlags(cmd *cobra.Command) {
	// Target selection
	cmd.Flags().Uint32Var(&flags.ID, "id", flags.ID,
		"eBPF program identifier to measure (typically the kernel bpf_prog id).")

	// Measurement window
	cmd.Flags().DurationVar(&flags.Duration, "duration", flags.Duration,
		"How long to collect samples for (e.g. 10s, 1m).")
	cmd.Flags().DurationVar(&flags.Warmup, "warmup", flags.Warmup,
		"Optional warmup period to discard before measurement (e.g. 5s).")

	// Stats config
	cmd.Flags().StringSliceVar(&flags.Percentiles, "percentiles", flags.Percentiles,
		"Percentile set to compute: default, wide, or tail. Example: --percentiles tail")

	// Output selection
	cmd.Flags().BoolVar(&flags.JSON, "json", flags.JSON,
		"If true, output results as JSON")
	cmd.Flags().BoolVar(&flags.Pretty, "pretty", flags.Pretty,
		"If true, pretty-print JSON output (only applies with --json).")
	cmd.Flags().StringVarP(&flags.Output, "output", "o", flags.Output,
		"Write output to a file instead of stdout.")

}
func (flags *LatencyFlags) ToOptions(parent string, args []string) (*LatencyOptions, error) {
	// Validation
	if flags.ID == 0 {
		return nil, fmt.Errorf("--id is required")
	}
	if flags.Duration == 0 {
		return nil, fmt.Errorf("--duration is required")
	}

	o := &LatencyOptions{
		ID:       flags.ID,
		Duration: flags.Duration,
	}

	// Handle optional warmup
	if flags.Warmup > 0 {
		o.Warmup = &flags.Warmup
	}

	// Determine output format
	if flags.JSON {
		o.Format = OutputJSON
	} else {
		o.Format = OutputText
	}
	o.Pretty = flags.Pretty

	// Handle output destination
	if flags.Output != "" {
		// Will be opened in Run()
		o.OutputPath = flags.Output
	}
	// Otherwise defaults to stdout in Run()

	// Parse percentiles (if specified)
	o.PercentileKeys = normalizePercentiles(flags.Percentiles)

	return o, nil
}

// normalizePercentiles converts user input to normalized keys
// e.g., ["50", "99.9"] -> ["p50", "p99_9"]
// e.g., ["default"] -> ["p50", "p90", "p99"]
// e.g., ["tail"] -> ["p90", "p99", "p99_9", "p99_99"]
func normalizePercentiles(input []string) []string {
	if len(input) == 0 {
		return []string{"p50", "p90", "p99", "p99_9"} // default
	}

	// Handle presets
	if len(input) == 1 {
		switch input[0] {
		case "default":
			return []string{"p50", "p90", "p99"}
		case "wide":
			// Generate p1, p2, p3, ..., p99, p100
			result := make([]string, 100)
			for i := 1; i <= 100; i++ {
				result[i-1] = fmt.Sprintf("p%d", i)
			}
			return result
		case "tail":
			return []string{"p90", "p95", "p99", "p99_9", "p99_99"}
		}
	}

	// Parse individual percentiles
	result := make([]string, 0, len(input))
	for _, p := range input {
		// Remove "p" prefix if present
		p = strings.TrimPrefix(p, "p")
		// Replace "." with "_"
		p = strings.ReplaceAll(p, ".", "_")
		result = append(result, "p"+p)
	}
	return result
}

func (o *LatencyOptions) Run() error {
	ctx := context.Background()

	// Setup output writer
	if err := o.setupOutput(); err != nil {
		return fmt.Errorf("setup output: %w", err)
	}
	defer o.closeOutput()

	// Create collector
	interval := 100 * time.Millisecond // sampling interval
	o.collector = collector.NewLatencyCollector(o.ID, interval, o.Warmup)

	// Start collector in background
	ctx, cancel := context.WithTimeout(ctx, o.Duration)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- o.collector.Start(ctx)
	}()

	// Live updates during measurement
	if o.Format == OutputText {
		if err := o.runWithLiveUpdates(ctx); err != nil {
			return err
		}
	} else {
		// JSON mode: just wait for completion
		if err := o.waitForCompletion(ctx, errCh); err != nil {
			return err
		}
	}

	// Output final statistics
	return o.outputFinalStats()
}

type LatencyOptions struct {

	// Target selection
	ID uint32

	// Measurement window
	Duration time.Duration
	Warmup   *time.Duration // nil => no warmup/discard

	// Output selection
	Format     OutputFormat
	Pretty     bool
	Out        io.Writer
	OutputPath string // file path (if specified)

	PercentileKeys []string // normalized: ["p50","p90","p99","p99_9"]

	// Internal (set during Run)
	collector *collector.LatencyCollector
}

type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

func (o *LatencyOptions) setupOutput() error {
	if o.OutputPath != "" {
		f, err := os.Create(o.OutputPath)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		o.Out = f
	} else {
		o.Out = os.Stdout
	}

	return nil
}
func (o *LatencyOptions) closeOutput() {
	if f, ok := o.Out.(*os.File); ok && f != os.Stdout {
		f.Close()
	}
}

func (o *LatencyOptions) runWithLiveUpdates(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second) // update every second
	defer ticker.Stop()

	fmt.Fprintf(o.Out, "Collecting latency stats for eBPF program %d...\n", o.ID)
	if o.Warmup != nil {
		fmt.Fprintf(o.Out, "Warmup period: %v\n", *o.Warmup)
	}
	fmt.Fprintf(o.Out, "Duration: %v\n\n", o.Duration)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Get current snapshot
			snapshot, err := o.collector.Snapshot()
			if err != nil {
				// No samples yet, skip
				continue
			}

			// Clear previous line (ANSI escape)
			fmt.Fprintf(o.Out, "\r\033[K")

			// Print live stats on same line
			latency := snapshot.(bpfsv1.Latency)
			fmt.Fprintf(o.Out, "Samples: %d | Mean: %v | StdDev: %v | Min: %v | Max: %v",
				latency.Samples,
				time.Duration(latency.Mean),
				time.Duration(latency.StdDev),
				time.Duration(*latency.Min),
				time.Duration(*latency.Max),
			)
		}
	}
}

func (o *LatencyOptions) waitForCompletion(ctx context.Context, errCh chan error) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		if err != nil && err != context.DeadlineExceeded {
			return err
		}
		return nil
	}
}

func (o *LatencyOptions) outputFinalStats() error {
	// Stop collector and get final snapshot
	if err := o.collector.Stop(); err != nil {
		return fmt.Errorf("stop collector: %w", err)
	}

	snapshot, err := o.collector.Snapshot()
	if err != nil {
		return fmt.Errorf("get final snapshot: %w", err)
	}

	// For text mode, add newline after live updates
	if o.Format == OutputText {
		fmt.Fprintln(o.Out, "\n\n=== Final Statistics ===")
	}

	var outputter output.ParameterOutput
	switch o.Format {
	case OutputText:
		outputter = &output.TextOutput{}
	case OutputJSON:
		outputter = &output.JsonOutput{}
	default:
		return fmt.Errorf("unknown output format: %v", o.Format)
	}

	// Use the output interface to format and write
	if err := outputter.OutputParam(snapshot, o.Out); err != nil {
		return fmt.Errorf("output statistics: %w", err)
	}

	return nil
}

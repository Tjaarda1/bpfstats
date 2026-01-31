/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
)

// latencyCmd represents the latency command
var latencyCmd = &cobra.Command{
	Use:   "latency",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("latency called")
	},
}

func init() {
	rootCmd.AddCommand(latencyCmd)

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

	o := &LatencyOptions{}

	return o, nil
}

func NewCmdLatency(parent string) *cobra.Command {
	flags := NewLatencyFlags()

	cmd := &cobra.Command{
		Use:                   "latency",
		DisableFlagsInUseLine: true,
		Short:                 "",
		Long:                  "",
		Example:               latencyExample,
		Run: func(cmd *cobra.Command, args []string) {
			o, _ := flags.ToOptions(parent, args)
			o.Run()
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

func (o *LatencyOptions) Run() error {

	return nil
}

type LatencyOptions struct {

	// Target selection
	ID uint32

	// Measurement window
	Duration time.Duration
	Warmup   *time.Duration // nil => no warmup/discard

	// Output selection
	Format OutputFormat
	Pretty bool
	Out    io.Writer

	PercentileKeys []string // normalized: ["p50","p90","p99","p99_9"]
}

type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

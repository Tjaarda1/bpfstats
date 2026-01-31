# bpfstat

bpfstat is a research-oriented command-line tool for precise statistical analysis of eBPF program execution.
It focuses on measuring and reporting per-program runtime, latency distributions, and CPU cost with well-defined semantics and controlled measurement error.

Unlike observability tools optimized for dashboards or coarse metrics, bpfstat is designed for methodologically sound performance analysis: reproducible measurements, explicit attribution, and statistically meaningful summaries (e.g., percentiles, variance, and tail behavior).

## Key goals

- Accurate per-invocation measurements of eBPF program execution time

- Distribution-aware statistics, including configurable percentiles, averages, and standard deviation

- Low and quantifiable instrumentation overhead, with explicit reporting of dropped or sampled events

- Clear measurement semantics, suitable for research, benchmarking, and performance evaluation

- Deterministic and reproducible output, designed for analysis rather than dashboards

# bpfstats

bpfstat is a research-oriented command-line tool for precise statistical analysis of eBPF program execution.
It focuses on measuring and reporting per-program runtime, latency distributions, and CPU cost with well-defined semantics and controlled measurement error.

Unlike observability tools optimized for dashboards or coarse metrics, bpfstat is designed for methodologically sound performance analysis: reproducible measurements, explicit attribution, and statistically meaningful summaries (e.g., percentiles, variance, and tail behavior).

## Key goals

- Accurate per-invocation measurements of eBPF program execution time

- Distribution-aware statistics, including configurable percentiles, averages, and standard deviation

- Low and quantifiable instrumentation overhead, with explicit reporting of dropped or sampled events

- Clear measurement semantics, suitable for research, benchmarking, and performance evaluation

- Deterministic and reproducible output, designed for analysis rather than dashboards

## Running

Enable BPF stats:

```bash
sudo sysctl -w kernel.bpf_stats_enabled=1
```

Check the BPF program ID:

```bash
sudo bpftool prog
```

Note the program ID you want to inspect.

Install bpftool if needed:

```bash
ARCH=$(dpkg --print-architecture)
VERSION=7.7.0

wget "https://github.com/libbpf/bpftool/releases/download/v${VERSION}/bpftool-v${VERSION}-${ARCH}.tar.gz"
tar -xvf "bpftool-v${VERSION}-${ARCH}.tar.gz" -C /usr/bin
```

Run the latency tool from this directory:

```bash
go run . latency --id <prog-id> --duration 60s
```

Example:

```bash
go run . latency --id 123 --duration 60s
```

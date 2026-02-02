package collector

import (
	"context"
	"time"

	bpfsv1 "github.com/Tjaarda1/bpfstats/api/v1"
)

type Collector interface {
	Start(ctx context.Context) error
	Stop() error
	Snapshot() (bpfsv1.Parameter, error)
}

// Optional: for bounded runs (CLI duration) with clear semantics.
type Windowed interface {
	// Reset clears previous state and sets the window semantics.
	Reset(window time.Duration, warmup *time.Duration)
	// Finalize freezes the stats (optional; Snapshot can also compute on demand).
	Finalize()
}

// func GetProgramId(name string) int64 {
//       p, err := ebpf.NewProgramFromID(ebpf.ProgramID(0))
//       if err != nil {
//               p = ebpf.ProgramGetNextID(id)
//               continue
//       }
//       if p.

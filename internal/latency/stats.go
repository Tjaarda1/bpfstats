package latency

import (
	"fmt"
	"log"
	"time"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

func GetLatency() {
	closer, err := ebpf.EnableStats(unix.BPF_STATS_RUN_TIME)
	if err != nil {
		log.Fatalf("EnableStats: %v", err)
	}
	defer closer.Close()

	//      GetProgramId("redirect_service")
	prog, err := ebpf.NewProgramFromID(ebpf.ProgramID(1015))
	if err != nil {
		log.Fatalf("NewProgramFromID: %v", err)
	}
	defer prog.Close()

	stats, err := prog.Stats()
	if err != nil {
		log.Fatalf("Stats: %v", err)
	}

	fmt.Printf("run_time_ns=%d\n", stats.Runtime)
	fmt.Printf("run_cnt=%d\n", stats.RunCount)

	if stats.RunCount > 0 {
		avg := time.Duration(int64(stats.Runtime) / int64(stats.RunCount))
		fmt.Printf("avg_per_call=%s\n", avg)
	}
}

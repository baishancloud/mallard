package sysprocfs

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/process"
)

var selfPid = int32(os.Getpid())

// ProcessStat is stats of current process,
// cpu, memory, connections
type ProcessStat struct {
	Memory                float64 `json:"memory"`
	MemoryPercent         float32 `json:"memory_percent"`
	CPUPercent            float64 `json:"cpu_percent"`
	Fds                   int32   `json:"fds"`
	Goroutines            int     `json:"goroutines"`
	Read, Write           int     `json:"read"`
	ReadBytes, WriteBytes int     `json:"read_bytes"`
}

// String prints memory friendly
func (k ProcessStat) String() string {
	s, _ := json.Marshal(k)
	return string(s)
}

var (
	oldTotal     float64 = -1
	oldTotalTime time.Time
	// cpus         = float64(runtime.NumCPU())
	oldStat *process.IOCountersStat
)

func getCPU(ps *process.Process) float64 {
	cpuStat, _ := ps.Times()
	if cpuStat != nil {
		total := cpuStat.Total()
		if oldTotal < 0 {
			oldTotal = total
			oldTotalTime = time.Now()
			return 0
		}
		totalDiff := total - oldTotal
		timeDiff := time.Since(oldTotalTime).Seconds()
		oldTotal = total
		oldTotalTime = time.Now()
		return totalDiff * 100 / timeDiff
	}
	return 0
}

// SelfProcessStat gets process stat of self program
func SelfProcessStat() (*ProcessStat, error) {
	p, err := process.NewProcess(selfPid)
	if err != nil {
		return nil, err
	}
	ps := new(ProcessStat)
	mem, _ := p.MemoryInfo()
	if mem != nil {
		ps.Memory = float64(mem.RSS) / 1024 / 1024
		ps.MemoryPercent, _ = p.MemoryPercent()
	}
	stat, err := p.IOCounters()
	if err == nil {
		if oldStat == nil {
			oldStat = stat
		} else {
			ps.Write = int(stat.WriteCount) - int(oldStat.WriteCount)
			ps.Read = int(stat.ReadCount) - int(oldStat.ReadCount)
			ps.WriteBytes = int(stat.WriteBytes) - int(oldStat.WriteBytes)
			ps.ReadBytes = int(stat.ReadBytes) - int(oldStat.ReadBytes)
			oldStat = stat
		}
	}
	ps.CPUPercent = getCPU(p)
	ps.Fds, _ = p.NumFDs()
	ps.Goroutines = runtime.NumGoroutine()
	return ps, nil
}

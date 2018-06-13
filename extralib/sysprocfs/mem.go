package sysprocfs

import (
	"encoding/json"

	"github.com/shirou/gopsutil/mem"
)

// MemoryStat is stats info of memory
type MemoryStat struct {
	Buffers      uint64 `json:"buffers"`
	Cached       uint64 `json:"cached"`
	Slab         uint64 `json:"slab"`
	Dirty        uint64 `json:"dirty"`
	Active       uint64 `json:"active"`
	Inactive     uint64 `json:"inactive"`
	Wired        uint64 `json:"wired"`
	Shared       uint64 `json:"shared"`
	Writeback    uint64 `json:"writeback"`
	WritebackTmp uint64 `json:"writeback_tmp"`
	PageTables   uint64 `json:"page_tables"`
	CommitLimit  uint64 `json:"commit_limit"`
	CommittedAS  uint64 `json:"committed_as"`

	MemTotal     uint64 `json:"mem_total"`
	MemFree      uint64 `json:"mem_free"`
	MemAvailable uint64 `json:"mem_available"` // for CentOS 7

	SwapTotal uint64 `json:"swap_total"`
	SwapUsed  uint64 `json:"swap_used"`
	SwapFree  uint64 `json:"swap_free"`
	SwapSin   uint64 `json:"swap_sin"`
	SwapSout  uint64 `json:"swap_sout"`

	AvailableRatio float64 `json:"available_ratio"`
	OccupyRatio    float64 `json:"occupy_ratio"`
	FreeRatio      float64 `json:"free_ratio"`
	UsedRatio      float64 `json:"used_ratio"`
	SwapFreeRatio  float64 `json:"swap_free_ratio"`
	SwapUsedRatio  float64 `json:"swap_used_ratio"`
}

// String prints memory friendly
func (m MemoryStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

// Memory return memory stat
func Memory() (*MemoryStat, error) {
	memData, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	swapData, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	m := &MemoryStat{
		Buffers:      memData.Buffers,
		Cached:       memData.Cached,
		Slab:         memData.Slab,
		Dirty:        memData.Dirty,
		Active:       memData.Active,
		Inactive:     memData.Inactive,
		Wired:        memData.Wired,
		Shared:       memData.Shared,
		MemTotal:     memData.Total,
		MemAvailable: memData.Available,
		Writeback:    memData.Writeback,
		WritebackTmp: memData.WritebackTmp,
		PageTables:   memData.PageTables,
		CommitLimit:  memData.CommitLimit,
		CommittedAS:  memData.CommittedAS,
		MemFree:      memData.Free,
		SwapTotal:    swapData.Total,
		SwapUsed:     swapData.Used,
		SwapFree:     swapData.Free,
		SwapSin:      swapData.Sin,
		SwapSout:     swapData.Sout,
	}
	var memAvailable uint64
	if m.MemAvailable > 0 {
		memAvailable = m.MemAvailable // MemAvailable is enabled in CentOS 7
	} else {
		memAvailable = m.MemFree + m.Buffers + m.Cached
	}
	memUsed := m.MemTotal - memAvailable

	pmemAvailable := 0.0
	pmemUsed := 0.0
	if m.MemTotal != 0 {
		pmemAvailable = float64(memAvailable) * 100.0 / float64(m.MemTotal)
		pmemUsed = float64(memUsed) * 100.0 / float64(m.MemTotal)
	}

	pswapFree := 0.0
	pswapUsed := 0.0
	if m.SwapTotal != 0 {
		pswapFree = float64(m.SwapFree) * 100.0 / float64(m.SwapTotal)
		pswapUsed = float64(m.SwapUsed) * 100.0 / float64(m.SwapTotal)
	}
	m.AvailableRatio = pmemAvailable
	m.FreeRatio = float64(m.MemFree) * 100.0 / float64(m.MemTotal)
	m.OccupyRatio = 100.0 - m.FreeRatio
	m.UsedRatio = pmemUsed
	m.SwapFreeRatio = pswapFree
	m.SwapUsedRatio = pswapUsed
	return m, nil
}

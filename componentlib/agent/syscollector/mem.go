package syscollector

import (
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	meminfoMetricName = "mem"
)

func init() {
	registerFactory("mem", MemoryMetrics)
}

// MemoryMetrics returns metric value of memory stat
func MemoryMetrics() ([]*models.Metric, error) {
	m, err := sysprocfs.Memory()
	if err != nil {
		return nil, err
	}

	pswapFree := 0.0
	pswapUsed := 0.0
	if m.SwapTotal != 0 {
		pswapFree = float64(m.SwapFree) * 100.0 / float64(m.SwapTotal)
		pswapUsed = float64(m.SwapUsed) * 100.0 / float64(m.SwapTotal)
	}

	var memAvailable uint64
	if m.MemAvailable > 0 {
		memAvailable = m.MemAvailable // MemAvailable is enabled in CentOS 7
	} else {
		memAvailable = m.MemFree + m.Buffers + m.Cached
	}
	memUsed := m.MemTotal - memAvailable
	pmemUsed := 0.0
	if m.MemTotal != 0 {
		pmemUsed = float64(memUsed) * 100.0 / float64(m.MemTotal)
	}

	mv := &models.Metric{
		Name:  meminfoMetricName,
		Value: float64(m.OccupyRatio),
		Fields: map[string]interface{}{
			"memtotal":     m.MemTotal,
			"memused":      memUsed,
			"memfree":      memAvailable,
			"memav":        memAvailable,
			"memfree0":     m.MemFree,
			"swaptotal":    m.SwapTotal,
			"swapused":     m.SwapUsed,
			"swapfree":     m.SwapFree,
			"memavperc":    utils.FixFloat(m.AvailableRatio),
			"memfreeperc":  utils.FixFloat(m.FreeRatio),
			"memusedperc":  utils.FixFloat(pmemUsed),
			"swapfreeperc": utils.FixFloat(pswapFree),
			"swapusedperc": utils.FixFloat(pswapUsed),
			"memcache":     m.Cached,
			"membuffer":    m.Buffers,
			"memslab":      m.Slab,
			"memdirty":     m.Dirty,
			"memactive":    m.Active,
			"meminactive":  m.Inactive,
			"memwired":     m.Wired,
			"memshared":    m.Shared,
			"pagetables":   m.PageTables,
			"writeback":    m.Writeback,
			"writebacktmp": m.WritebackTmp,
			"commitlimit":  m.CommitLimit,
			"committedas":  m.CommittedAS,
		},
	}

	return []*models.Metric{mv}, nil
}

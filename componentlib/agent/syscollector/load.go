package syscollector

import (
	"runtime"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	loadAvgMetricName  = "load"
	loadMiscMetricName = "load_procs"
)

func init() {
	registerFactory("load", LoadMetrics)
}

// LoadMetrics returns load data metrics
func LoadMetrics() ([]*models.Metric, error) {
	avgM, err := sysprocfs.LoadAvg()
	if err != nil {
		return nil, err
	}
	misc, err := sysprocfs.LoadMisc()
	if err != nil {
		return nil, err

	}
	core := float64(runtime.NumCPU())
	serious := utils.FixFloat(avgM.Load1 / core)
	m := &models.Metric{
		Name:  loadAvgMetricName,
		Value: utils.FixFloat(avgM.Load1),
		Fields: map[string]interface{}{
			"1min":           avgM.Load1,
			"5min":           avgM.Load5,
			"15min":          avgM.Load15,
			"cpu_core_count": core,
			"serious":        serious,
		},
	}
	m3 := &models.Metric{
		Name:  loadMiscMetricName,
		Value: float64(misc.ProcsRunning),
		Fields: map[string]interface{}{
			"procs_running": misc.ProcsRunning,
			"procs_blocked": misc.ProcsBlocked,
			"ctxt":          misc.Ctxt,
			"ctxt_diff":     misc.CtxtDiff,
		},
	}
	return []*models.Metric{m, m3}, nil
}

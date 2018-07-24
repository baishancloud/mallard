package syscollector

import (
	"runtime"
	"strings"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

func init() {
	registerFactory("core.cpu", CPUAllMetrics)
}

const (
	cpuMetricName         = "cpu"
	cpuCoreMetricName     = "cpu_core"
	cpuCoreMockMetricName = "cpu_core_mock"
)

// CPUAllMetrics return each and total cpu cores metric values
func CPUAllMetrics() ([]*models.Metric, error) {
	metrics := make([]*models.Metric, 0, runtime.NumCPU()+2)
	total, err := CPUTotalMetrics()
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, total...)
	cores, err := CPUCoresMetrics()
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, cores...)
	return metrics, nil
}

// CPUCoresMetrics return metrics of each core usages
func CPUCoresMetrics() ([]*models.Metric, error) {
	coresM, err := sysprocfs.CPUCores()
	if err != nil {
		return nil, err
	}
	if len(coresM) == 0 {
		return nil, nil
	}
	var metrics []*models.Metric
	var over95Count int
	for _, cpuM := range coresM {
		fieldsMap := map[string]interface{}{
			"idle":       utils.FixFloat(cpuM.Idle),
			"usr":        utils.FixFloat(cpuM.User),
			"nice":       utils.FixFloat(cpuM.Nice),
			"sys":        utils.FixFloat(cpuM.System),
			"iowait":     utils.FixFloat(cpuM.Iowait),
			"irq":        utils.FixFloat(cpuM.Irq),
			"softirq":    utils.FixFloat(cpuM.Softirq),
			"steal":      utils.FixFloat(cpuM.Steal),
			"guest":      utils.FixFloat(cpuM.Guest),
			"use":        utils.FixFloat(100 - cpuM.Idle),
			"guest_nice": utils.FixFloat(cpuM.GuestNice),
			"stolen":     utils.FixFloat(cpuM.Stolen),
		}
		if cpuM.Idle < 5 {
			over95Count++
		}
		m := &models.Metric{
			Name:   cpuCoreMetricName,
			Value:  utils.FixFloat(cpuM.Idle),
			Fields: fieldsMap,
			Tags: map[string]string{
				"core": strings.TrimPrefix(cpuM.CPU, "cpu"),
			},
		}
		metrics = append(metrics, m)
	}
	over95Percent := utils.FixFloat(float64(over95Count) / float64(len(coresM)) * 100)
	m95 := &models.Metric{
		Name:  cpuCoreMockMetricName,
		Value: over95Percent,
		Fields: map[string]interface{}{
			"rate":      over95Percent,
			"count":     over95Count,
			"valueslen": len(coresM),
		},
		Tags: map[string]string{
			"user": "core95",
		},
	}
	metrics = append(metrics, m95)
	return metrics, nil
}

// CPUTotalMetrics return metric of cpu total
func CPUTotalMetrics() ([]*models.Metric, error) {
	cpuM, err := sysprocfs.CPUTotal()
	if err != nil {
		return nil, err
	}
	if cpuM == nil {
		return nil, nil
	}

	fieldsMap := map[string]interface{}{
		"idle":       utils.FixFloat(cpuM.Idle),
		"usr":        utils.FixFloat(cpuM.User),
		"nice":       utils.FixFloat(cpuM.Nice),
		"sys":        utils.FixFloat(cpuM.System),
		"iowait":     utils.FixFloat(cpuM.Iowait),
		"irq":        utils.FixFloat(cpuM.Irq),
		"softirq":    utils.FixFloat(cpuM.Softirq),
		"steal":      utils.FixFloat(cpuM.Steal),
		"guest":      utils.FixFloat(cpuM.Guest),
		"use":        utils.FixFloat(100 - cpuM.Idle),
		"guest_nice": utils.FixFloat(cpuM.GuestNice),
		"stolen":     utils.FixFloat(cpuM.Stolen),
	}
	m := &models.Metric{
		Name:   cpuMetricName,
		Value:  utils.FixFloat(cpuM.Idle),
		Fields: fieldsMap,
	}
	return []*models.Metric{m}, nil
}

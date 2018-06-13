package syscollector

import (
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	kernelMetricName          = "kernel"
	fileDescriptionMetricName = "file_description"
)

func init() {
	registerFactory("kernel", KernelMetrics)
}

// KernelMetrics returns metric value of file descriptors and processes
func KernelMetrics() ([]*models.Metric, error) {
	knStat, err := sysprocfs.Kernel()
	if err != nil {
		return nil, err
	}
	m := &models.Metric{
		Name:  kernelMetricName,
		Value: float64(knStat.FilesAllocated),
		Fields: map[string]interface{}{
			"maxfiles":       knStat.FilesMax,
			"maxproc":        knStat.ProcessMax,
			"filesallocated": knStat.FilesAllocated,
		},
	}
	m2 := &models.Metric{
		Name:  fileDescriptionMetricName,
		Value: float64(knStat.FilesAllocated),
	}
	return []*models.Metric{m, m2}, nil
}

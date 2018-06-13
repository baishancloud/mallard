package syscollector

import (
	"math"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	diskBytesMetricName  = "df_bytes"
	diskInodesMetricName = "df_inodes"
)

func init() {
	registerFactory("disk.usage", DiskUsagesMetrics)
}

// DiskUsagesMetrics collects mounted devices usage and inode usage metric value
func DiskUsagesMetrics() ([]*models.Metric, error) {
	usages, err := sysprocfs.DiskUsages()
	if err != nil {
		return nil, err
	}

	var (
		diskTotal uint64
		diskUsed  uint64
		metrics   = make([]*models.Metric, 0, len(usages)+1)
	)

	for _, use := range usages {
		if use.Usage == nil {
			continue
		}

		diskTotal += use.Usage.Total
		diskUsed += use.Usage.Used

		tags := map[string]string{
			"mount":  use.PartitionStat.Mountpoint,
			"fstype": use.PartitionStat.Fstype,
		}

		// use df percent, used/(used+free), not free/total,
		usedPercent := float64(use.Usage.Used) / float64(use.Usage.Used+use.Usage.Free) * 100
		if math.IsNaN(usedPercent) {
			usedPercent = 0
		}
		m := &models.Metric{
			Name:  diskBytesMetricName,
			Value: float64(use.Usage.Total),
			Fields: map[string]interface{}{
				"total":       use.Usage.Total,
				"used":        use.Usage.Used,
				"free":        use.Usage.Free,
				"usedpercent": utils.FixFloat(usedPercent),
				"freepercent": utils.FixFloat(100 - usedPercent),
			},
			Tags: tags,
		}
		m2 := &models.Metric{
			Name:  diskInodesMetricName,
			Value: float64(use.Usage.InodesTotal),
			Fields: map[string]interface{}{
				"total":       use.Usage.InodesTotal,
				"used":        use.Usage.InodesUsed,
				"free":        use.Usage.InodesFree,
				"usedpercent": utils.FixFloat(use.Usage.InodesUsedPercent),
				"freepercent": utils.FixFloat(100 - use.Usage.InodesUsedPercent),
			},
			Tags: tags,
		}
		metrics = append(metrics, m, m2)
	}
	return metrics, nil
}

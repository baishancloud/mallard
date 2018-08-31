package syscollector

import (
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	diskIOMetricName     = "disk_io"
	diskIOStatMetricName = "disk_io_stat"
	diskIOMaxMetricName  = "disk_io_maxutil"
)

func init() {
	registerFactory("disk.io", DiskIOMetrics)
	registerFactory("disk.iostat", IOStatsMetrics)
}

// DiskIOMetrics returns disk io values metric values
func DiskIOMetrics() ([]*models.Metric, error) {
	ioMap, err := sysprocfs.DiskIO()
	if err != nil {
		return nil, err
	}

	metrics := make([]*models.Metric, 0, len(ioMap))

	for _, ds := range ioMap {
		m := &models.Metric{
			Name:  diskIOMetricName,
			Value: float64(ds.ReadCount),
			Fields: map[string]interface{}{
				"read_requests":       ds.ReadCount,
				"read_merged":         ds.MergedReadCount,
				"read_bytes":          ds.ReadBytes,
				"msec_read":           ds.ReadTime,
				"write_requests":      ds.WriteCount,
				"write_merged":        ds.MergedWriteCount,
				"write_bytes":         ds.WriteBytes,
				"msec_write":          ds.WriteTime,
				"ios_in_progress":     ds.IopsInProgress,
				"msec_total":          ds.IoTime,
				"msec_weighted_total": ds.WeightedIO,
			},
			Tags: map[string]string{
				"device": ds.Name,
				"mount":  ds.Mount,
				// "serial": ds.SerialNumber,
			},
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// IOStatsMetrics is io stats data to metric values
func IOStatsMetrics() ([]*models.Metric, error) {
	ioStats, err := sysprocfs.IOStats()
	if err != nil || len(ioStats) == 0 {
		return nil, err
	}

	var (
		metrics = make([]*models.Metric, 0, len(ioStats)+1)
		maxUtil = -1.0
	)

	for _, io := range ioStats {
		m := &models.Metric{
			Name:  diskIOStatMetricName,
			Value: io.Await,
			Fields: map[string]interface{}{
				"read_bytes":         float64(io.ReadBytes),
				"write_bytes":        float64(io.WriteBytes),
				"read_count":         io.ReadCount,
				"read_merged_count":  io.MergedReadCount,
				"write_count":        io.WriteCount,
				"write_merged_count": io.MergedWriteCount,
				"avgrq_sz":           utils.FixFloat(io.AvgrqSz),
				"avgqu_sz":           utils.FixFloat(io.AvgquSz),
				"await":              utils.FixFloat(io.Await),
				"r_await":            utils.FixFloat(io.RAwait),
				"w_await":            utils.FixFloat(io.WAwait),
				"svctm":              utils.FixFloat(io.Svctm),
				"util":               utils.FixFloat(io.Util),
			},
			Tags: map[string]string{
				"device": io.Device,
			},
		}
		metrics = append(metrics, m)

		if io.Util > maxUtil {
			maxUtil = io.Util
		}
	}
	if maxUtil >= 0 {
		m := &models.Metric{
			Name:  diskIOMaxMetricName,
			Value: float64(int(maxUtil)),
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

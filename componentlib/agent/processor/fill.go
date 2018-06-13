package processor

import (
	"time"

	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/corelib/models"
)

// FillMetrics fills agent info to metrics, if fields are empty
func FillMetrics(metrics []*models.Metric) []*models.Metric {
	real := make([]*models.Metric, 0, len(metrics))
	now := time.Now().Unix()
	for i := range metrics {
		m := metrics[i]
		if m == nil {
			log.Warn("nil-metric")
			continue
		}
		if m.Name == "" {
			log.Warn("no-name-endpoint", "metric", m)
			continue
		}
		if m.Endpoint == "" {
			m.Endpoint = serverinfo.Hostname()
		}
		if m.Time == 0 {
			m.Time = now
		}
		m.FillTags(serverinfo.Sertypes(),
			serverinfo.Cachegroup(),
			serverinfo.StorageGroup(),
			serverinfo.Hostname())
		real = append(real, m)
	}
	return real
}

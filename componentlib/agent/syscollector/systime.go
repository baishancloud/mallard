package syscollector

import (
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

func init() {
	registerFactory("systime", SystimeMetrics)
}

var (
	transferTime int64
	localTime    int64
	sysTimeDiff  int64
)

const (
	systimeMetricName = "systime"
)

// SetSystime set system time
func SetSystime(t int64) {
	if t == 0 {
		transferTime = 0
		localTime = 0
		sysTimeDiff = 0
		return
	}
	transferTime = t
	localTime = time.Now().Unix()
	sysTimeDiff = transferTime - localTime
	if sysTimeDiff < 0 {
		sysTimeDiff = 0 - sysTimeDiff
	}
}

// SystimeMetrics get system time metric
func SystimeMetrics() ([]*models.Metric, error) {
	return []*models.Metric{
		{
			Name:  systimeMetricName,
			Value: float64(sysTimeDiff),
			Fields: map[string]interface{}{
				"portal":  transferTime,
				"local":   localTime,
				"absdiff": sysTimeDiff,
			},
		},
	}, nil
}

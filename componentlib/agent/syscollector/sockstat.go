package syscollector

import (
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	netSockstatMetricName = "sockstat"
)

func init() {
	registerFactory("net.sockstat", SockstatMetrics)
}

// SockstatMetrics is metric value of /proc/net/sockstat
func SockstatMetrics() ([]*models.Metric, error) {
	socks, err := sysprocfs.Sockstat()
	if err != nil {
		return nil, err
	}
	m := &models.Metric{
		Name:   netSockstatMetricName,
		Value:  float64(socks["sockets_used"]),
		Fields: map[string]interface{}{},
	}
	for k, v := range socks {
		m.Fields[k] = v
	}
	return []*models.Metric{m}, nil
}

package syscollector

import (
	"strconv"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

var (
	// Ports is port list for detecting, such as [8080,8081,443]
	Ports []int64
)

const (
	portMetricName = "net_port_listen"
)

func init() {
	registerFactory("net.port", PortMetrics)
}

// PortMetrics returns metric value of port listening stats
func PortMetrics() ([]*models.Metric, error) {
	if len(Ports) == 0 {
		return nil, nil
	}
	allListeningPorts, err := sysprocfs.TCPPorts()
	if err != nil {
		return nil, err
	}
	var metrics []*models.Metric
	for _, p := range Ports {
		isFound := false
		for _, listening := range allListeningPorts {
			if p == listening {
				isFound = true
			}
		}
		m := &models.Metric{
			Name:  portMetricName,
			Value: 0,
			Tags: map[string]string{
				"port": strconv.FormatInt(p, 10),
			},
		}
		if isFound {
			m.Value = 1
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

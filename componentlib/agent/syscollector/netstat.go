package syscollector

import (
	"errors"
	"strings"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	netStatTCPExMetricName = "tcp_ext"
	netConnectionsName     = "netstat_connections"
)

func init() {
	registerFactory("net.tcpext", NetStatTCPExMetrics)
	// registerFactory("net.connetions", NetConnectionsMetrics)
}

// NetStatTCPExMetrics is metric value of TcpEx in netstat
func NetStatTCPExMetrics() ([]*models.Metric, error) {
	netStats, err := sysprocfs.NetStat()
	if err != nil {
		return nil, err
	}
	tcpExts := netStats["TcpExt"]
	if len(tcpExts) == 0 {
		return nil, errors.New("no-TcpExt")
	}
	m := &models.Metric{
		Name:   netStatTCPExMetricName,
		Value:  float64(tcpExts["TCPLoss"]),
		Fields: make(map[string]interface{}),
	}
	for key, val := range tcpExts {
		m.Fields[key] = val
	}
	return []*models.Metric{m}, nil
}

// NetConnectionsMetrics is metric value of net connection counters
func NetConnectionsMetrics() ([]*models.Metric, error) {
	counters, err := sysprocfs.NetConnections()
	if err != nil {
		return nil, err
	}
	m := &models.Metric{
		Name:   netConnectionsName,
		Value:  float64(counters["ESTABLISHED"]),
		Fields: map[string]interface{}{},
	}
	for k, c := range counters {
		m.Fields[strings.ToLower(k)] = c
	}
	return []*models.Metric{m}, nil
}

package syscollector

import (
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

func init() {
	registerFactory("snmp.tcp", SnmpTCPMetrics)
	registerFactory("snmp.udp", SnmpUDPMetrics)
}

const (
	udpMetricName = "snmp_udp"
)

// SnmpUDPMetrics returns metric value of udp stats from snmp
func SnmpUDPMetrics() ([]*models.Metric, error) {
	udp, err := sysprocfs.SnmpUDP()
	if err != nil {
		return nil, err
	}
	m := &models.Metric{
		Name:   udpMetricName,
		Value:  float64(udp["InDatagrams"]),
		Fields: make(map[string]interface{}, len(udp)),
	}
	for k, v := range udp {
		m.Fields[k] = v
	}
	return []*models.Metric{m}, nil
}

const (
	tcpMetricName       = "snmp_tcp"
	tcpPluginMetricName = "tcp_plugin"
)

// SnmpTCPMetrics returns metric value of tcp stats from snmp
func SnmpTCPMetrics() ([]*models.Metric, error) {
	tcpM, err := sysprocfs.SnmpTCP()
	if err != nil {
		return nil, err
	}
	if len(tcpM) == 0 {
		return nil, nil
	}
	m := &models.Metric{
		Name:   tcpMetricName,
		Value:  tcpM["CurrEstab"],
		Fields: map[string]interface{}{},
	}
	for k, v := range tcpM {
		m.Fields[k] = v
	}
	m2 := &models.Metric{
		Name:  tcpPluginMetricName,
		Value: utils.FixFloat(tcpM["Retrans"]),
		Fields: map[string]interface{}{
			"retrans_rate": utils.FixFloat(tcpM["Retrans"]),
			"conn":         m.Value,
		},
	}
	return []*models.Metric{m, m2}, nil
}

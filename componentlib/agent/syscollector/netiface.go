package syscollector

import (
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	netIfaceMetricName         = "netif"
	netIfacePackagesMetricName = "netif_packages"
)

func init() {
	registerFactory("net.iface", NetIfaceMetrics)

	// init first value
	sysprocfs.NetIfaceStats()
}

// NetIfaceMetrics returns net stats of one iface card
func NetIfaceMetrics() ([]*models.Metric, error) {
	ifacesMap, err := sysprocfs.NetIfaceStats()
	if err != nil {
		return nil, err
	}
	ret := make([]*models.Metric, 0, len(ifacesMap))
	for iface, netIf := range ifacesMap {
		m2 := &models.Metric{
			Name:   netIfaceMetricName,
			Value:  float64(netIf["tx_bytes"] + netIf["rx_bytes"]),
			Fields: make(map[string]interface{}, len(netIf)),
			Tags: map[string]string{
				"iface": iface,
			},
		}
		for k, v := range netIf {
			m2.Fields[k] = v
		}
		m2.Fields["total_bytes"] = m2.Value
		m2.Fields["total_packets"] = float64(netIf["tx_packets"] + netIf["rx_packets"])
		m2.Fields["total_errors"] = float64(netIf["tx_errors"] + netIf["rx_errors"])
		ret = append(ret, m2)
	}
	return ret, nil
}

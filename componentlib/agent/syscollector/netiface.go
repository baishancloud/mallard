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
}

// NetIfaceMetrics returns net stats of one iface card
func NetIfaceMetrics() ([]*models.Metric, error) {
	ifacesMap, err := sysprocfs.NetIfaceStats()
	if err != nil {
		return nil, err
	}
	ret := make([]*models.Metric, 0, len(ifacesMap))
	for _, netIf := range ifacesMap {
		m := &models.Metric{
			Name:  netIfaceMetricName,
			Value: float64(netIf.OutBytes),
			Fields: map[string]interface{}{
				"inbytes":         netIf.InBytes,
				"inpackets":       netIf.InPackages,
				"inerrors":        netIf.InErrors,
				"indropped":       netIf.InDropped,
				"infifoerrs":      netIf.InFifoErrs,
				"outbytes":        netIf.OutBytes,
				"outpackets":      netIf.OutPackages,
				"outerrors":       netIf.OutErrors,
				"outdropped":      netIf.OutDropped,
				"outfifoerrs":     netIf.OutFifoErrs,
				"totalbytes":      netIf.TotalBytes,
				"totalpackets":    netIf.TotalPackages,
				"totalerrors":     netIf.TotalErrors,
				"totaldropped":    netIf.TotalDropped,
				"in_bandwidth":    netIf.InBandwidth,
				"out_bandwidth":   netIf.OutBandwidth,
				"total_bandwidth": netIf.TotalBandwidth,
			},
			Tags: map[string]string{
				"iface": netIf.Iface,
			},
		}
		ret = append(ret, m)

		if len(netIf.Packages) > 0 {
			m2 := &models.Metric{
				Name:   netIfacePackagesMetricName,
				Value:  float64(netIf.Packages["tx_bytes"] + netIf.Packages["rx_bytes"]),
				Fields: make(map[string]interface{}, len(netIf.Packages)),
				Tags: map[string]string{
					"iface": netIf.Iface,
				},
			}
			for k, v := range netIf.Packages {
				m2.Fields[k] = v
			}
			ret = append(ret, m2)
		}
	}
	return ret, nil
}

package syscollector

import (
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

const (
	netIfaceMetricName = "netif"
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
	}
	return ret, nil
}

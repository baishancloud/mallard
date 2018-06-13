package sysprocfs

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/shirou/gopsutil/net"
)

var (
	lastNetIfaceStats     = make(map[string]*NetIfaceStat)
	lastNetIfaceStatsLock sync.RWMutex
)

// NetIfaceStat is stats for one net interface
type NetIfaceStat struct {
	Iface          string  `json:"iface"`
	InBytes        uint64  `json:"in_bytes"`
	InPackages     uint64  `json:"in_packages"`
	InErrors       uint64  `json:"in_errors"`
	InDropped      uint64  `json:"in_dropped"`
	InFifoErrs     uint64  `json:"in_fifo_errs"`
	OutBytes       uint64  `json:"out_bytes"`
	OutPackages    uint64  `json:"out_packages"`
	OutErrors      uint64  `json:"out_errors"`
	OutDropped     uint64  `json:"out_dropped"`
	OutFifoErrs    uint64  `json:"out_fifo_errs"`
	TotalBytes     uint64  `json:"total_bytes"`
	TotalPackages  uint64  `json:"total_packages"`
	TotalErrors    uint64  `json:"total_errors"`
	TotalDropped   uint64  `json:"total_dropped"`
	Time           int64   `json:"time"`
	InBandwidth    float64 `json:"in_bandwidth"`
	OutBandwidth   float64 `json:"out_bandwidth"`
	TotalBandwidth float64 `json:"total_bandwidth"`
}

// String prints memory friendly
func (n NetIfaceStat) String() string {
	s, _ := json.Marshal(n)
	return string(s)
}

// NetIfaceStats return net interfaces stats
// if set prefix, it only returns prefixed named net interfaces
// otherwise, return fall
func NetIfaceStats() (map[string]*NetIfaceStat, error) {
	lastNetIfaceStatsLock.Lock()
	defer lastNetIfaceStatsLock.Unlock()

	ifaceList, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}
	rawMap := make(map[string]net.IOCountersStat)
	for i := range ifaceList {
		rawMap[ifaceList[i].Name] = ifaceList[i]
	}

	now := time.Now().Unix()
	resultMap := make(map[string]*NetIfaceStat, len(rawMap))
	for k, info := range rawMap {
		iface := &NetIfaceStat{
			Iface:         info.Name,
			InBytes:       info.BytesRecv,
			OutBytes:      info.BytesSent,
			InDropped:     info.Dropin,
			OutDropped:    info.Dropout,
			InErrors:      info.Errin,
			OutErrors:     info.Errout,
			InFifoErrs:    info.Fifoin,
			OutFifoErrs:   info.Fifoout,
			InPackages:    info.PacketsRecv,
			OutPackages:   info.PacketsSent,
			TotalBytes:    info.BytesRecv + info.BytesSent,
			TotalDropped:  info.Dropin + info.Dropout,
			TotalErrors:   info.Errin + info.Errout,
			TotalPackages: info.PacketsRecv + info.PacketsSent,
			Time:          now,
		}
		if iface.TotalBytes < 1 {
			continue
		}
		last := lastNetIfaceStats[k]
		if last == nil {
			lastNetIfaceStats[k] = iface
		} else {
			totalBytesDiff := iface.TotalBytes - last.TotalBytes
			deltaTime := float64(iface.Time - last.Time)
			iface.InBandwidth = float64(iface.InBytes-last.InBytes) / deltaTime
			iface.OutBandwidth = float64(iface.OutBytes-last.OutBytes) / deltaTime
			iface.TotalBandwidth = float64(totalBytesDiff) / deltaTime
		}
		resultMap[k] = iface
	}
	return resultMap, nil
}

// NetInterfaces return netcard interfaces info
func NetInterfaces() ([]net.InterfaceStat, error) {
	return net.Interfaces()
}

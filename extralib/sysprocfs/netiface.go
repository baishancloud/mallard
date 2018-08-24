package sysprocfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/net"
)

var (
	netIfacePackagesFile  = "/sys/class/net/%s/statistics"
	lastNetIfaceStats     = make(map[string]*NetIfaceStat)
	lastNetIfaceStatsLock sync.RWMutex
)

// NetIfaceStat is stats for one net interface
type NetIfaceStat struct {
	Iface          string           `json:"iface"`
	InBytes        uint64           `json:"in_bytes"`
	InPackages     uint64           `json:"in_packages"`
	InErrors       uint64           `json:"in_errors"`
	InDropped      uint64           `json:"in_dropped"`
	InFifoErrs     uint64           `json:"in_fifo_errs"`
	OutBytes       uint64           `json:"out_bytes"`
	OutPackages    uint64           `json:"out_packages"`
	OutErrors      uint64           `json:"out_errors"`
	OutDropped     uint64           `json:"out_dropped"`
	OutFifoErrs    uint64           `json:"out_fifo_errs"`
	TotalBytes     uint64           `json:"total_bytes"`
	TotalPackages  uint64           `json:"total_packages"`
	TotalErrors    uint64           `json:"total_errors"`
	TotalDropped   uint64           `json:"total_dropped"`
	Time           int64            `json:"time"`
	InBandwidth    float64          `json:"in_bandwidth"`
	OutBandwidth   float64          `json:"out_bandwidth"`
	TotalBandwidth float64          `json:"total_bandwidth"`
	Packages       map[string]int64 `json:"packages"`

	rawPackages map[string]int64
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
			rawPackages:   netIfacePackages(info.Name),
		}
		if iface.TotalBytes < 1 {
			continue
		}
		last := lastNetIfaceStats[k]
		if last == nil {
			lastNetIfaceStats[k] = iface
		} else {
			// calculate bandwidth
			totalBytesDiff := iface.TotalBytes - last.TotalBytes
			deltaTime := float64(iface.Time - last.Time)
			iface.InBandwidth = float64(iface.InBytes-last.InBytes) / deltaTime
			iface.OutBandwidth = float64(iface.OutBytes-last.OutBytes) / deltaTime
			iface.TotalBandwidth = float64(totalBytesDiff) / deltaTime

			// calculate packages
			if len(iface.rawPackages) > 0 && len(last.rawPackages) > 0 {
				iface.Packages = make(map[string]int64)
				for name, value := range iface.rawPackages {
					value2, ok := last.rawPackages[name]
					if ok {
						diff := value - value2
						if diff < 0 {
							diff = 0
						}
						iface.Packages[name] = diff
					}
				}
			}
		}
		resultMap[k] = iface
	}
	return resultMap, nil
}

func netIfacePackages(iface string) map[string]int64 {
	dir := fmt.Sprintf(netIfacePackagesFile, iface)
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	result := make(map[string]int64, 15)
	filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		raw, _ := ioutil.ReadFile(fpath)
		if len(raw) > 0 {
			value, err := strconv.ParseInt(strings.Trim(string(raw), "\n"), 10, 64)
			if err == nil {
				basename := filepath.Base(fpath)
				result[basename] = value
			}
		}
		return nil
	})
	return result
}

// NetInterfaces return netcard interfaces info
func NetInterfaces() ([]net.InterfaceStat, error) {
	return net.Interfaces()
}

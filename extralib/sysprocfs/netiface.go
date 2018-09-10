package sysprocfs

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/net"
)

var (
	// netIfacePackagesFile = "/sys/class/net/%s/statistics"
	netDevFile = "/proc/net/dev"
	netDevKeys = map[int]string{
		0:  "rx_bytes",
		1:  "rx_packets",
		2:  "rx_errors",
		3:  "rx_dropped",
		4:  "rx_fifo_errors",
		5:  "rx_frame_errors",
		6:  "rx_compressed",
		7:  "multicast",
		8:  "tx_bytes",
		9:  "tx_packets",
		10: "tx_errors",
		11: "tx_dropped",
		12: "tx_fifo_errors",
		13: "collisions",
		14: "tx_carrier_errors",
		15: "tx_compressed",
	}
	lastNetIfaceStats     = make(map[string]map[string]int64)
	lastNetIfaceStatsLock sync.RWMutex
)

// NetDevStats returns counter from /proc/net/dev, raw data
func NetDevStats() (map[string]map[string]int64, error) {
	fh, err := os.Open(netDevFile)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	ret := make(map[string]map[string]int64)
	reader := bufio.NewReader(fh)
	for {
		var bs []byte
		bs, err = readLine(reader)
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return nil, err
		}

		line := string(bs)
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		ifaceName := strings.TrimSpace(line[:idx])
		fields := strings.Fields(line[idx+1:])
		if len(fields) != 16 {
			continue
		}
		values := make(map[string]int64)
		for idx, field := range fields {
			key := netDevKeys[idx]
			values[key], err = strconv.ParseInt(field, 10, 64)
			if err != nil {
				return nil, err
			}
		}
		if values["tx_bytes"]+values["rx_bytes"] == 0 {
			continue
		}
		ret[ifaceName] = values
	}
	return ret, nil
}

// NetIfaceStats return net interfaces stats
func NetIfaceStats() (map[string]map[string]int64, error) {
	lastNetIfaceStatsLock.Lock()
	defer lastNetIfaceStatsLock.Unlock()

	stats, err := NetDevStats()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]map[string]int64)
	for iface, values := range stats {
		last := lastNetIfaceStats[iface]
		if len(last) == 0 {
			lastNetIfaceStats[iface] = values
			continue
		}
		results := make(map[string]int64, len(values))
		for key, value := range values {
			value2, ok := last[key]
			if ok {
				diff := value - value2
				if diff < 0 {
					diff = 0
				}
				results[key] = diff
			}
		}
		lastNetIfaceStats[iface] = values
		ret[iface] = results
	}
	return ret, nil
}

// NetInterfaces return netcard interfaces info
func NetInterfaces() ([]net.InterfaceStat, error) {
	return net.Interfaces()
}

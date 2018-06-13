package sysprocfs

import "github.com/shirou/gopsutil/net"

var (
	lastSNMPTCPData *net.ProtoCountersStat
)

// SnmpTCP returns tcp stats from snmp
// it records data at first run, then calculate and return delta data
func SnmpTCP() (map[string]float64, error) {
	tcpData, err := SNMP("tcp")
	if err != nil {
		return nil, err
	}
	if lastSNMPTCPData == nil {
		lastSNMPTCPData = tcpData
		return nil, nil
	}
	fields := make(map[string]float64)
	for key, value := range tcpData.Stats {
		if key == "CurrEstab" || key == "MaxConn" {
			fields[key] = float64(value)
			continue
		}
		fields[key] = float64(value)
		if key == "RetransSegs" || key == "OutSegs" {
			fields[key+"Diff"] = float64(tcpData.Stats[key] - lastSNMPTCPData.Stats[key])
		}
	}
	if fields["OutSegsDiff"] > 0 {
		fields["Retrans"] = fields["RetransSegsDiff"] / fields["OutSegsDiff"] * 100
	} else {
		fields["Retrans"] = 0
	}
	lastSNMPTCPData = tcpData
	return fields, nil
}

// SnmpUDP return udp stats from snmp
func SnmpUDP() (map[string]int64, error) {
	udpData, err := SNMP("udp")
	if err != nil || udpData == nil {
		return nil, err
	}
	return udpData.Stats, nil
}

// SNMP return general stat from snmp
func SNMP(title string) (*net.ProtoCountersStat, error) {
	counters, err := net.ProtoCounters([]string{title})
	if err != nil {
		return nil, err
	}
	for _, c := range counters {
		if c.Protocol == title {
			return &c, nil
		}
	}
	return nil, nil
}

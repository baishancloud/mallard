package sysprocfs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/shirou/gopsutil/net"
)

// NetStat return TcpExt and IpExt from netstat
func NetStat() (map[string]map[string]uint64, error) {
	fs, err := os.Open("/proc/net/netstat")
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	ret := make(map[string]map[string]uint64)
	reader := bufio.NewReader(fs)
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

		title := strings.TrimSpace(line[:idx])
		stat := make(map[string]uint64)
		ths := strings.Fields(strings.TrimSpace(line[idx+1:]))
		// the next line must be values
		bs, err = readLine(reader)
		if err != nil {
			return nil, err
		}
		valLine := string(bs)
		tds := strings.Fields(strings.TrimSpace(valLine[idx+1:]))
		for i := 0; i < len(ths); i++ {
			stat[ths[i]], err = strconv.ParseUint(tds[i], 10, 64)
			if err != nil {
				return nil, err
			}
		}
		ret[title] = stat
	}
	return ret, nil
}

// NetConnections return connection stat counts
func NetConnections() (map[string]int, error) {
	netconns, err := net.Connections("all")
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, netcon := range netconns {
		if netcon.Type == syscall.SOCK_DGRAM {
			counts["UDP"]++
			continue
		}
		counts[netcon.Status]++
	}
	return counts, nil
}

func runCmdBytes(name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.Bytes(), err
}

// TCPPorts return tcp listening ports
func TCPPorts() ([]int64, error) {
	return listeningPorts("sh", "-c", "ss -t -l -n")
}

// UDPPorts return udp listening ports
func UDPPorts() ([]int64, error) {
	return listeningPorts("sh", "-c", "ss -u -a -n")
}

func listeningPorts(name string, args ...string) ([]int64, error) {
	ports := []int64{}

	bs, err := runCmdBytes(name, args...)
	if err != nil {
		return ports, err
	}

	reader := bufio.NewReader(bytes.NewBuffer(bs))

	line, err := readLine(reader)
	if err != nil {
		return ports, err
	}
	for {
		line, err = readLine(reader)
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return ports, err
		}
		fields := strings.Fields(string(line))
		fieldsLen := len(fields)
		if fieldsLen != 4 && fieldsLen != 5 {
			return ports, fmt.Errorf("output of %s format not supported", name)
		}
		portColumnIndex := 2
		if fieldsLen == 5 {
			portColumnIndex = 3
		}

		location := strings.LastIndex(fields[portColumnIndex], ":")
		port := fields[portColumnIndex][location+1:]

		p, e := strconv.ParseInt(port, 10, 64)
		if e != nil {
			return ports, fmt.Errorf("parse port to int64 fail: %s", e.Error())
		}
		ports = append(ports, p)
	}
	return utils.Int64SliceUnique(ports), nil
}

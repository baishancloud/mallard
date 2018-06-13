package sysprocfs

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"
)

var (
	sockstatSockets = []byte("sockets:")
	sockstatTCP     = []byte("TCP:")
	sockstatUDP     = []byte("UDP:")
	sockstatColon   = []byte(":")
)

func lineToMap(line []byte, prefix string) map[string]int64 {
	fields := bytes.Fields(line)
	if len(fields)%2 != 0 {
		return nil
	}
	m := make(map[string]int64)
	for k, v := range fields {
		if k%2 == 1 {
			m[prefix+"_"+string(fields[k-1])], _ = strconv.ParseInt(string(v), 10, 64)
		}
	}
	return m
}

func readLine(r *bufio.Reader) ([]byte, error) {
	line, isPrefix, err := r.ReadLine()
	for isPrefix && err == nil {
		var bs []byte
		bs, isPrefix, err = r.ReadLine()
		line = append(line, bs...)
	}
	return line, err
}

// Sockstat returns sock stats from /proc/net/sockstat
func Sockstat() (map[string]int64, error) {
	fs, err := os.Open("/proc/net/sockstat")
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	reader := bufio.NewReader(fs)
	m := make(map[string]int64)
	for {
		line, err := readLine(reader)
		if bytes.HasPrefix(line, sockstatSockets) ||
			bytes.HasPrefix(line, sockstatTCP) ||
			bytes.HasPrefix(line, sockstatUDP) {
			idx := bytes.Index(line, sockstatColon)
			line2 := line[idx+1:]
			values := lineToMap(line2, strings.ToLower(strings.TrimSpace(string(line[:idx]))))
			for k, v := range values {
				m[k] = v
			}
		}
		if err != nil {
			break
		}
	}
	return m, nil
}

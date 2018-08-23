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
	sockstatTCP6    = []byte("TCP6:")
	sockstatUDP     = []byte("UDP:")
	sockstatUDP6    = []byte("UDP6:")
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
	return sockStatGet("/proc/net/sockstat")
}

// Sockstat6 returns sock stats from /proc/net/sockstat6
func Sockstat6() (map[string]int64, error) {
	return sockStatGet("/proc/net/sockstat6")
}

func sockStatGet(file string) (map[string]int64, error) {
	fs, err := os.Open(file)
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
			bytes.HasPrefix(line, sockstatTCP6) ||
			bytes.HasPrefix(line, sockstatUDP) ||
			bytes.HasPrefix(line, sockstatUDP6) {
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

package sysprocfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/host"
)

// KernelStat is stat of kernel info
type KernelStat struct {
	FilesMax       int64          `json:"files_max"`
	ProcessMax     int64          `json:"process_max"`
	FilesAllocated int64          `json:"files_allocated"`
	Host           *host.InfoStat `json:"host"`
}

// String prints memory friendly
func (k KernelStat) String() string {
	s, _ := json.Marshal(k)
	return string(s)
}

// Kernel gets all kernel info
func Kernel() (*KernelStat, error) {
	kr := &KernelStat{}
	var err error
	if kr.FilesMax, err = FilesMax(); err != nil {
		return nil, err
	}
	if kr.ProcessMax, err = ProcessMax(); err != nil {
		return nil, err
	}
	if kr.FilesAllocated, err = FilesAllocated(); err != nil {
		return nil, err
	}
	if kr.Host, err = HostInfo(); err != nil {
		return nil, err
	}
	return kr, nil
}

// FilesMax get max open files in system
func FilesMax() (int64, error) {
	bytes, err := ioutil.ReadFile("/proc/sys/fs/file-max")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
}

// ProcessMax get max open thread process in system
func ProcessMax() (int64, error) {
	bytes, err := ioutil.ReadFile("/proc/sys/kernel/pid_max")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
}

// FilesAllocated get opend files in system
func FilesAllocated() (int64, error) {
	bytes, err := ioutil.ReadFile("/proc/sys/fs/file-nr")
	if err != nil {
		return 0, err
	}
	arr := strings.Fields(strings.TrimSpace(string(bytes)))
	if len(arr) != 3 {
		return 0, fmt.Errorf("format error")
	}
	return strconv.ParseInt(arr[0], 10, 64)
}

// HostInfo return host information
func HostInfo() (*host.InfoStat, error) {
	return host.Info()
}

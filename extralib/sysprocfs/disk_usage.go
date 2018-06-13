package sysprocfs

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/disk"
)

// MountsFilter is default filter to strip mounts
var MountsFilter = map[string][]string{
	"device": {
		"none",
		"nodev",
	},
	"fstype": {
		"cgroup",
		"debugfs",
		"devpts",
		"devtmpfs",
		"rpc_pipefs",
		"rootfs",
	},
	"mount": {
		"/sys",
		"/net",
		"/misc",
		"/proc",
		"/lib",
	},
}

// DiskMountUsage contains mount stat and usage stats for one disk device
type DiskMountUsage struct {
	PartitionStat *disk.PartitionStat `json:"partition_stat,omitempty"`
	Usage         *disk.UsageStat     `json:"usage,omitempty"`
}

// String prints memory friendly
func (d DiskMountUsage) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}

// DiskMounts return mounts with default filter
func DiskMounts() (map[string]*disk.PartitionStat, error) {
	mounts, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*disk.PartitionStat)
	for i, m := range mounts {
		var isIgnore bool
		for _, device := range MountsFilter["device"] {
			if m.Device == device {
				isIgnore = true
				continue
			}
		}
		if isIgnore {
			continue
		}

		for _, fstype := range MountsFilter["fstype"] {
			if m.Fstype == fstype {
				isIgnore = true
				continue
			}
			if strings.HasPrefix(m.Fstype, "fuse") {
				isIgnore = true
				continue
			}
		}
		if isIgnore {
			continue
		}

		for _, mount := range MountsFilter["mount"] {
			if strings.HasPrefix(m.Mountpoint, mount) {
				isIgnore = true
				continue
			}
		}
		if isIgnore {
			continue
		}
		result[m.Mountpoint] = &mounts[i]
	}
	return result, nil
}

// DiskUsage return disk spaces and inodes usages by mount point
func DiskUsage(mount string) (*disk.UsageStat, error) {
	return disk.Usage(mount)
}

func cleanMounts(mounts map[string]*disk.PartitionStat) map[string]*disk.PartitionStat {
	devicesMount := make(map[string][]string)
	for _, mount := range mounts {
		devicesMount[mount.Device] = append(devicesMount[mount.Device], mount.Mountpoint)
	}
	cleans := make(map[string]bool)
	for _, mountpoints := range devicesMount {
		if len(mountpoints) < 2 {
			continue
		}
		sort.Sort(sort.StringSlice(mountpoints))
		for i, dir := range mountpoints {
			for j, dir2 := range mountpoints {
				if i == j {
					continue
				}
				if strings.HasPrefix(dir2, dir) {
					cleans[dir2] = true
				}
			}
		}
	}
	for key, mount := range mounts {
		if cleans[mount.Mountpoint] {
			delete(mounts, key)
		}
	}
	return mounts
}

// DiskUsages return disk spaces and inodes usages with default mounts list
func DiskUsages() (map[string]*DiskMountUsage, error) {
	mounts, err := DiskMounts()
	if err != nil {
		return nil, fmt.Errorf("mount error %s", err.Error())
	}

	mounts = cleanMounts(mounts)

	result := make(map[string]*DiskMountUsage)
	for k, v := range mounts {
		use, err := DiskUsage(v.Mountpoint)
		if err != nil {
			continue
		}
		result[k] = &DiskMountUsage{
			PartitionStat: mounts[k],
			Usage:         use,
		}
	}
	return result, nil
}

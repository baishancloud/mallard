package sysprocfs

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/disk"
)

const (
	diskSectorSize float64 = 512
)

// DiskIOStat is disk io stats
type DiskIOStat struct {
	disk.IOCountersStat
	Mount string
	Time  time.Time
}

// DiskIO return disk io stats
func DiskIO() (map[string]*DiskIOStat, error) {
	diskIO, err := disk.IOCounters()
	if err != nil {
		return nil, err
	}
	mounts, err := DiskMounts()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	result := make(map[string]*DiskIOStat)
	for k, v := range diskIO {
		if ShouldDiskIODevice(v.Name) {
			result[k] = &DiskIOStat{
				IOCountersStat: diskIO[k],
				Time:           now,
			}
			for _, m := range mounts {
				if strings.HasPrefix(m.Device, k) {
					result[k].Mount = m.Mountpoint
				}
			}
		}
	}
	return result, nil
}

// ShouldDiskIODevice returns is disk should detected io
func ShouldDiskIODevice(device string) bool {
	normal := len(device) == 3 && (strings.HasPrefix(device, "sd") || strings.HasPrefix(device, "vd")) || strings.HasPrefix(device, "disk")
	aws := len(device) >= 4 && strings.HasPrefix(device, "xvd")
	return normal || aws
}

var (
	lastDiskMap  = make(map[string]*DiskIOStat)
	lastDiskLock sync.Mutex
)

// IOStat is io stats for one disk in duration
type IOStat struct {
	ReadBytes        float64 `json:"read_bytes"`
	WriteBytes       float64 `json:"write_bytes"`
	ReadCount        uint64  `json:"read_count"`
	MergedReadCount  uint64  `json:"merged_read_count"`
	WriteCount       uint64  `json:"write_count"`
	MergedWriteCount uint64  `json:"merged_write_count"`

	AvgrqSz    float64 `json:"avgrq_sz"`
	AvgquSz    float64 `json:"avgqu_sz"`
	Await      float64 `json:"await"`
	RAwait     float64 `json:"r_await"`
	WAwait     float64 `json:"w_await"`
	Svctm      float64 `json:"svctm"`
	Util       float64 `json:"util"`
	Device     string  `json:"device"`
	Rio        float64 `json:"rio"`
	Wio        float64 `json:"wio"`
	TotalIo    uint64  `json:"total_io"`
	DeltaResc  float64 `json:"delta_resc"`
	DeltaWsec  float64 `json:"delta_wsec"`
	DeltaTotal float64 `json:"delta_total"`
}

// String prints memory friendly
func (k IOStat) String() string {
	s, _ := json.Marshal(k)
	return string(s)
}

// IOStats return io stats for all disks in duration
// it compares to last record and calculates values with time duration
func IOStats() (map[string]*IOStat, error) {
	diskIO, err := DiskIO()
	if err != nil {
		return nil, err
	}
	lastDiskLock.Lock()
	defer lastDiskLock.Unlock()
	if len(lastDiskMap) == 0 {
		lastDiskMap = diskIO
		return nil, nil
	}

	result := make(map[string]*IOStat)

	for k, ds := range diskIO {
		if lastDiskMap[k] == nil {
			continue
		}
		l := lastDiskMap[k]
		rio := float64(ds.ReadCount) - float64(l.ReadCount)
		if rio < 0 {
			rio = 0
		}
		wio := float64(ds.WriteCount) - float64(l.WriteCount)
		if wio < 0 {
			wio = 0
		}
		deltaRsec := (float64(ds.ReadBytes) - float64(l.ReadBytes)) / diskSectorSize
		if deltaRsec < 0 {
			deltaRsec = 0
		}
		deltaWsec := (float64(ds.WriteBytes) - float64(l.WriteBytes)) / diskSectorSize
		if deltaWsec < 0 {
			deltaWsec = 0
		}
		ruse := float64(ds.ReadTime) - float64(l.ReadTime)
		if ruse < 0 {
			ruse = 0
		}
		wuse := float64(ds.WriteTime) - float64(l.WriteTime)
		if wuse < 0 {
			wuse = 0
		}
		use := float64(ds.IoTime) - float64(l.IoTime)
		if use < 0 {
			use = 0
		}
		nio := rio + wio
		if nio < 0 {
			nio = 0
		}
		var (
			avgrqSz = 0.0
			await   = 0.0
			rAwait  = 0.0
			wAwait  = 0.0
			svctm   = 0.0
		)
		if nio > 0 {
			avgrqSz = float64(deltaRsec+deltaWsec) / float64(nio)
			await = float64(ruse+wuse) / float64(nio)
			svctm = float64(use) / float64(nio)
		}
		if wio > 0 {
			wAwait = float64(wuse) / float64(wio)
		}
		if rio > 0 {
			rAwait = float64(ruse) / float64(rio)
		}
		avgquSz := float64(ds.WeightedIO-l.WeightedIO) / 1000

		duration := ds.Time.Sub(l.Time).Seconds() * 1000
		util := float64(use) * 100.0 / float64(duration)
		if util > 100.0 {
			util = 100.0
		}

		ioS := &IOStat{
			ReadBytes:        deltaRsec * diskSectorSize,
			WriteBytes:       deltaWsec * diskSectorSize,
			ReadCount:        ds.ReadCount - l.ReadCount,
			MergedReadCount:  ds.MergedReadCount - l.MergedReadCount,
			WriteCount:       ds.WriteCount - l.WriteCount,
			MergedWriteCount: ds.MergedWriteCount - l.MergedWriteCount,
			AvgrqSz:          avgrqSz,
			AvgquSz:          avgquSz,
			Await:            await,
			RAwait:           rAwait,
			WAwait:           wAwait,
			Svctm:            svctm,
			Util:             util,
			Device:           ds.Name,
			Rio:              rio,
			Wio:              wio,
			DeltaResc:        deltaRsec,
			DeltaWsec:        deltaWsec,
			DeltaTotal:       deltaRsec + deltaWsec,
		}
		result[k] = ioS
	}
	lastDiskMap = diskIO
	return result, nil
}

package sysprocfs

import (
	"sync"

	"github.com/shirou/gopsutil/cpu"
)

var (
	lastCPUTotalData *cpu.TimesStat
	lastCPUCoresData = make(map[string]*cpu.TimesStat)
	lastCPUCoresLock sync.Mutex
)

// CPUInfo return cpu processor info
func CPUInfo() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

// CPUCores return each cpu core times as map
func CPUCores() (map[string]*cpu.TimesStat, error) {
	coresData, err := cpu.Times(true)
	if err != nil {
		return nil, err
	}

	lastCPUCoresLock.Lock()
	result := make(map[string]*cpu.TimesStat)
	for i, c := range coresData {
		l := lastCPUCoresData[c.CPU]
		if l == nil {
			lastCPUCoresData[c.CPU] = &coresData[i]
			continue
		}
		delTotal := c.Total() - l.Total()
		one := &cpu.TimesStat{
			CPU:       c.CPU,
			Idle:      (c.Idle - l.Idle) * 100 / delTotal,
			User:      (c.User - l.User) * 100 / delTotal,
			Nice:      (c.Nice - l.Nice) * 100 / delTotal,
			System:    (c.System - l.System) * 100 / delTotal,
			Iowait:    (c.Iowait - l.Iowait) * 100 / delTotal,
			Irq:       (c.Irq - l.Irq) * 100 / delTotal,
			Softirq:   (c.Softirq - l.Softirq) * 100 / delTotal,
			Steal:     (c.Steal - l.Steal) * 100 / delTotal,
			Guest:     (c.Guest - l.Guest) * 100 / delTotal,
			GuestNice: (c.GuestNice - l.GuestNice) * 100 / delTotal,
			Stolen:    (c.Stolen - l.Stolen) * 100 / delTotal,
		}
		lastCPUCoresData[c.CPU] = &coresData[i]
		result[c.CPU] = one
	}
	lastCPUCoresLock.Unlock()

	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

const cpuTotalName = "cpu-total"

// CPUTotal return total-cpu times usages
// When first call, record data,
// Then call CPUTotal to get delta values
func CPUTotal() (*cpu.TimesStat, error) {
	cpuData, err := cpu.Times(false)
	if err != nil {
		return nil, err
	}
	for _, c := range cpuData {
		if c.CPU != cpuTotalName {
			continue
		}
		if lastCPUTotalData == nil {
			lastCPUTotalData = &c
			return nil, nil
		}
		l := lastCPUTotalData
		delTotal := c.Total() - l.Total()
		result := &cpu.TimesStat{
			CPU:       c.CPU,
			Idle:      (c.Idle - l.Idle) * 100 / delTotal,
			User:      (c.User - l.User) * 100 / delTotal,
			Nice:      (c.Nice - l.Nice) * 100 / delTotal,
			System:    (c.System - l.System) * 100 / delTotal,
			Iowait:    (c.Iowait - l.Iowait) * 100 / delTotal,
			Irq:       (c.Irq - l.Irq) * 100 / delTotal,
			Softirq:   (c.Softirq - l.Softirq) * 100 / delTotal,
			Steal:     (c.Steal - l.Steal) * 100 / delTotal,
			Guest:     (c.Guest - l.Guest) * 100 / delTotal,
			GuestNice: (c.GuestNice - l.GuestNice) * 100 / delTotal,
			Stolen:    (c.Stolen - l.Stolen) * 100 / delTotal,
		}
		lastCPUTotalData = &c
		return result, nil
	}
	return nil, nil
}

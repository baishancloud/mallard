package sysprocfs

import (
	"encoding/json"
	"sync"

	"github.com/shirou/gopsutil/load"
)

// LoadAvg return load data
func LoadAvg() (*load.AvgStat, error) {
	return load.Avg()
}

var (
	lastLoadMisc     *MiscStat
	lastLoadMiscLock sync.Mutex
)

// MiscStat extends load.MiscStat with ctxt difference value
type MiscStat struct {
	*load.MiscStat
	CtxtDiff int `json:"ctxt_diff"`
}

// String overwrites load.MiscStat.String
func (m MiscStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

// LoadMisc return process running stat
func LoadMisc() (*MiscStat, error) {
	current, err := load.Misc()
	if err != nil {
		return nil, err
	}

	lastLoadMiscLock.Lock()
	defer lastLoadMiscLock.Unlock()

	if lastLoadMisc == nil {
		lastLoadMisc = &MiscStat{
			MiscStat: current,
			CtxtDiff: 0,
		}
		return lastLoadMisc, nil
	}
	mt := &MiscStat{
		MiscStat: current,
		CtxtDiff: current.Ctxt - lastLoadMisc.Ctxt,
	}
	lastLoadMisc = mt
	return lastLoadMisc, nil
}

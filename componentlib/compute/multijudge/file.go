package multijudge

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

var (
	strategyFile    = "multi_strategies.log"
	strategyFileMod int64
)

// ScanStrategies scans strategies from file
func ScanStrategies(file string) {
	if file == "" {
		log.Warn("no-strategy-file")
		return
	}
	log.Debug("strategy-file", "file", file)
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		readStrategiesFile(file)
		<-ticker.C
	}
}

func readStrategiesFile(file string) {
	info, err := os.Stat(file)
	if err != nil {
		log.Warn("stat-strategy-file-error", "error", err)
		return
	}
	modTime := info.ModTime().Unix()
	if modTime == strategyFileMod {
		return
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Warn("read-strategy-file-error", "error", err)
		return
	}
	ss := make(map[int]*MultiStrategy)
	if err = json.Unmarshal(b, &ss); err != nil {
		log.Warn("decode-strategy-file-error", "error", err)
		return
	}
	SetStrategies(ss)
	strategyFileMod = modTime
	log.Info("read-strategy-file", "strategies", len(ss))
}

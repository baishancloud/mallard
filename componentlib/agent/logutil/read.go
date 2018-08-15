package logutil

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

var (
	readDir   string
	files     = make(map[string]int64)
	filesLock sync.RWMutex

	readCount = expvar.NewDiff("logtool.read")
)

func init() {
	expvar.Register(readCount)
}

// SetReadDir sets reading directory
func SetReadDir(dir string) {
	if dir == "" {
		return
	}
	os.MkdirAll(dir, os.ModePerm)
	readDir = dir
	log.Info("init-read", "dir", dir)
}

func readDirMetrics() ([]*models.Metric, error) {
	if readDir == "" {
		return nil, nil
	}
	var metrics []*models.Metric
	err := filepath.Walk(readDir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(fpath) == ".log" {
			if time.Since(info.ModTime()).Seconds() > 3600 {
				return nil
			}
			ms, err := readFile(fpath)
			if err != nil {
				log.Warn("read-error", "error", err, "file", fpath)
				return nil
			}
			if len(ms) > 0 {
				record := make(map[string]struct{})
				for _, m := range ms {
					record[m.Name] = struct{}{}
				}
				log.Debug("read", "file", fpath, "len", len(ms), "names", utils.KeysOfStructMap(record))
				metrics = append(metrics, ms...)
			}
		}
		return nil
	})
	return metrics, err
}

func readFile(file string) ([]*models.Metric, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	// check mod time
	modTime := info.ModTime().UnixNano()
	filesLock.Lock()
	defer filesLock.Unlock()
	isSame := (modTime == files[file])
	if isSame {
		return nil, nil
	}
	files[file] = modTime
	// read file
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var ms []*models.Metric
	if err = json.Unmarshal(bytes, &ms); err != nil {
		return nil, err
	}
	return ms, nil
}

// ReadInterval reads interval
func ReadInterval(interval time.Duration, ch chan []*models.Metric) {
	if readDir == "" {
		return
	}
	utils.Ticker(interval, readOnce(ch))
}

func readOnce(ch chan []*models.Metric) func() {
	return func() {
		metrics, err := readDirMetrics()
		if err != nil {
			log.Warn("read-error", "error", err)
		} else {
			mLen := int64(len(metrics))
			if mLen > 0 {
				log.Info("read", "len", mLen)
				if ch != nil {
					ch <- metrics
				}
				readCount.Incr(mLen)
			}
		}
	}
}

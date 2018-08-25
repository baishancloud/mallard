package judgestore

import (
	"os"
	"path/filepath"
	"time"

	"github.com/baishancloud/mallard/corelib/utils"
)

// DefaultExpireDuration is default metric values expire time in seconds
const DefaultExpireDuration = 10

// RunClean runs cleaning tasks
func RunClean() {
	if writingDir == "" {
		return
	}
	go autoSync()
	go autoClose()
	go autoClean()
}

func autoSync() {
	utils.TickerThen(time.Second*10, syncFileHandlers)
}

func syncFileHandlers() {
	writingFilesLock.Lock()
	for _, fn := range writingFiles {
		if fn == nil {
			continue
		}
		fn.Sync()
	}
	writingFilesLock.Unlock()
}

func closeFileHandlers() {
	writingFilesLock.Lock()
	var count int
	for key, fn := range writingFiles {
		if fn == nil {
			continue
		}
		duration := fn.IdleDuration()
		expire := DefaultExpireDuration
		filterLock.RLock()
		if filter := filters[fn.Metric()]; filter != nil && filter.Expire > 0 {
			expire = filter.Expire
		}
		filterLock.RUnlock()
		if duration.Minutes() > float64(expire) {
			fn.Close()
			delete(writingFiles, key)
			os.RemoveAll(fn.Name())
			log.Debug("fs-close", "name", fn.Name(), "du", int(duration.Seconds()), "expire", expire*60)
			count++
		}
	}
	log.Info("fs-remove-ok", "count", count)
	writingFilesLock.Unlock()
}

func autoClose() {
	time.Sleep(time.Second * 3)
	utils.TickerThen(time.Minute, closeFileHandlers)
}

const (
	// MaxLogTime is max log time duration of each metric log file since last modified time
	MaxLogTime = time.Minute * 30
)

func cleanFiles() {
	var size int64
	var count int64
	err := filepath.Walk(writingDir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(fpath) != ".log" {
			return nil
		}
		if time.Since(info.ModTime()) > MaxLogTime {
			os.RemoveAll(fpath)
			log.Debug("file-remove", "fpath", fpath)
			return nil
		}
		count++
		size += info.Size()
		return nil
	})
	if err != nil {
		log.Warn("clean-error", "error", err)
		return
	}
	log.Info("clean-ok", "file_mb", utils.FixFloat(float64(size)/1024/1024), "files", count)
	StoreSizeCounter.Set(int64(size))
	StoreFilesCounter.Set(count)
}

func autoClean() {
	if writingDir == "" {
		return
	}
	time.Sleep(time.Second * 5) // separate time with autoClose
	utils.Ticker(time.Minute, cleanFiles)
}

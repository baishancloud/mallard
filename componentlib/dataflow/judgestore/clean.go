package judgestore

import (
	"os"
	"path/filepath"
	"time"

	"github.com/baishancloud/mallard/corelib/utils"
)

// DefaultExpireDuration is default metric values expire time in minutes
const DefaultExpireDuration = 10

// RunClean runs cleaning tasks
func RunClean() {
	if writingDir == "" {
		return
	}
	go autoClose()
	go autoClean()
}

func autoClose() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		now := <-ticker.C
		writingFilesLock.Lock()
		var count int64
		for key, fn := range writingFiles {
			if fn == nil {
				continue
			}
			fn.Sync()
			if now.Unix()%20 != 0 {
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
		if now.Unix()%20 == 0 {
			log.Info("fs-remove-ok", "count", count)
		}
		writingFilesLock.Unlock()
	}
}

const (
	// MaxLogTime is max log time duration of each metric log file since last modified time
	MaxLogTime = time.Minute * 30
)

func autoClean() {
	if writingDir == "" {
		return
	}
	time.Sleep(time.Second * 5) // separate time with autoClose
	for {
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
			time.Sleep(time.Minute)
			continue
		}
		log.Info("clean-ok", "file_mb", utils.FixFloat(float64(size)/1024/1024), "files", count)
		StoreSizeCounter.Set(int64(size))
		StoreFilesCounter.Set(count)
		time.Sleep(time.Minute)
	}
}

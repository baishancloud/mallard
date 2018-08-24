package queues

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/baishancloud/mallard/corelib/container"
)

// Queue is common queue to handle packets data
type Queue struct {
	queue     container.LimitedQueue
	dumpDir   string
	queueSize int
}

// Push pushes raw pack to queue
func (q *Queue) Push(raw Packet) (int, bool) {
	if !q.queue.Push(raw) {
		_, count, err := q.Dump(q.queueSize / 5)
		if err != nil {
			return 0, false
		}
		return count, false
	}
	return 0, true
}

// NewQueue creates new queue with size and dump dir
func NewQueue(size int, dumpDir string) *Queue {
	if dumpDir != "" {
		os.MkdirAll(dumpDir, os.ModePerm)
	}
	return &Queue{
		queue:     container.NewLimitedList(size),
		queueSize: size,
		dumpDir:   dumpDir,
	}
}

// Pop pop items from item and encode to bytes with json and gzip
func (q *Queue) Pop(size int) (Packets, error) {
	data := q.queue.PopBatch(size)
	if len(data) == 0 {
		return nil, nil
	}
	dataLen := int64(len(data))
	result := make(Packets, 0, dataLen)
	for _, v := range data {
		if p, ok := v.(Packet); ok {
			result = append(result, p)
		}
	}
	return result, nil
}

// Len returns queue's current length
func (q *Queue) Len() int {
	return q.queue.Len()
}

var (
	// ErrDumpDirBlank means dumpdir is not set
	ErrDumpDirBlank = errors.New("dump-dir-blank")
)

func dumpFileName(dir string) string {
	return filepath.Join(dir, fmt.Sprintf("dump_%s.pack", time.Now().Format("20060102150405")))
}

// Dump dumps queue value of size
func (q *Queue) Dump(size int) (string, int, error) {
	if q.dumpDir == "" {
		return "", 0, ErrDumpDirBlank
	}
	values := q.queue.PopBatch(size)
	if len(values) == 0 {
		return "", 0, nil
	}
	fname := dumpFileName(q.dumpDir)
	os.MkdirAll(filepath.Dir(fname), os.ModePerm)
	b, err := json.Marshal(values)
	if err != nil {
		return "", 0, err
	}
	return fname, len(values), ioutil.WriteFile(fname, b, os.ModePerm)
}

// ReadLatestDump reads latest dump file to queue
func (q *Queue) ReadLatestDump() (string, int, error) {
	if q.dumpDir == "" {
		return "", 0, ErrDumpDirBlank
	}
	latest := ""
	var latestT time.Time
	err := filepath.Walk(q.dumpDir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(fpath) != ".pack" {
			return nil
		}
		if latest == "" {
			latest = fpath
			latestT = info.ModTime()
			return nil
		}
		if info.ModTime().Unix() > latestT.Unix() {
			latest = fpath
			latestT = info.ModTime()
			return nil
		}
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	if latest == "" {
		return "", 0, nil
	}
	fh, err := os.Open(latest)
	if err != nil {
		return "", 0, err
	}
	decoder := json.NewDecoder(fh)
	data := make(Packets, 0, q.queueSize/10)
	if err = decoder.Decode(&data); err != nil {
		fh.Close()
		return "", 0, err
	}
	fh.Close()

	var count int
	for _, item := range data {
		ok := q.queue.Push(item)
		if !ok {
			break
		}
		count++
	}
	return latest, count, os.Remove(latest)
}

// ScanDumpResult is result of once scan and reading dump file
type ScanDumpResult struct {
	Error error  `json:"error,omitempty"`
	Count int    `json:"count,omitempty"`
	File  string `json:"file,omitempty"`
}

// ScanDump scans dumped file and loads it
func (q *Queue) ScanDump(interval time.Duration, fn func(ScanDumpResult)) {
	for {
		file, count, err := q.ReadLatestDump()
		if file == "" && count == 0 && err == nil {
			time.Sleep(interval)
			continue
		}
		if fn != nil {
			fn(ScanDumpResult{
				File:  file,
				Count: count,
				Error: err,
			})
		}
		time.Sleep(interval)
	}
}

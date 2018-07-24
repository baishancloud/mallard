package judgestore

import (
	"os"
	"time"
)

type fileHandler struct {
	metric     string
	name       string
	fh         *os.File
	writeTime  time.Time
	createTime time.Time
}

func newFileHandler(name, metric string) (*fileHandler, error) {
	fh, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &fileHandler{
		metric:     metric,
		name:       name,
		fh:         fh,
		createTime: time.Now(),
	}, nil
}

func (fh *fileHandler) File() *os.File {
	return fh.fh
}

func (fh *fileHandler) Metric() string {
	return fh.metric
}

func (fh *fileHandler) Name() string {
	return fh.name
}

func (fh *fileHandler) Touch() {
	fh.writeTime = time.Now()
}

func (fh *fileHandler) Sync() error {
	return fh.fh.Sync()
}

func (fh *fileHandler) Close() error {
	return fh.fh.Close()
}

func (fh *fileHandler) IdleDuration() time.Duration {
	return time.Since(fh.writeTime)
}

func (fh *fileHandler) Duration() time.Duration {
	return time.Since(fh.createTime)
}

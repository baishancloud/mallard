package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	// ErrNoFileInfo means fileinfo is nil
	ErrNoFileInfo = errors.New("no-fileinfo")
	// ErrWrongFilename means wrong filename
	ErrWrongFilename = errors.New("wrong-filename")
)

// Plugin is executor of a plugin file
type Plugin struct {
	File         string
	LogFile      string
	FileModTime  int64
	LastExecTime int64
	Cycle        int64
	ReloadTime   int64
	timeout      time.Duration
}

// NewPlugin new plugin with file
func NewPlugin(file string, logFile string, info os.FileInfo) (*Plugin, error) {
	cycle, err := parseFilename(filepath.Base(file))
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, ErrNoFileInfo
	}
	p := &Plugin{
		File:         file,
		LogFile:      logFile,
		FileModTime:  info.ModTime().Unix(),
		Cycle:        cycle,
		LastExecTime: time.Now().Unix() - cycle + rand.Int63n(60), // set last exec time as an old time, so all plugins can run in different tick
		timeout:      calTimeout(cycle),
	}
	return p, nil
}

func parseFilename(file string) (int64, error) {
	idx := strings.Index(file, "_")
	if idx < 0 {
		return 0, ErrWrongFilename
	}
	return strconv.ParseInt(file[:idx], 10, 64)
}

func calTimeout(cycle int64) time.Duration {
	if cycle < 30 {
		return time.Second * 30
	}
	return time.Duration(cycle-1) * time.Second
}

// ShouldExec check plugin's execution time is arrived
func (p *Plugin) ShouldExec(t int64) bool {
	return t-p.LastExecTime >= p.Cycle
}

var (
	asNewBytesPrefix = []byte("!!")
)

// Exec execute plugin file
func (p *Plugin) Exec() ([]*models.Metric, error) {
	p.LastExecTime = time.Now().Unix()

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	var (
		stdout = bytes.NewBuffer(nil)
		stderr = bytes.NewBuffer(nil)
	)
	cmd := exec.CommandContext(ctx, p.File)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	if stderr.Len() > 0 {
		p.writeLog(stderr.Bytes())
	}

	if err != nil {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, err
	}

	if stdout.Len() == 0 {
		return nil, nil
	}
	outBytes := stdout.Bytes()
	if bytes.HasPrefix(outBytes, asNewBytesPrefix) {
		var metrics []*models.Metric
		outBytes = bytes.TrimPrefix(outBytes, asNewBytesPrefix) // trim prefix bytes
		if err = json.Unmarshal(outBytes, &metrics); err != nil {
			p.writeLog(outBytes)
			return nil, err
		}
		return metrics, nil
	}

	var metricsOld []*models.MetricRaw
	if err = json.Unmarshal(outBytes, &metricsOld); err != nil {
		p.writeLog(outBytes)
		return nil, err
	}
	if len(metricsOld) == 0 {
		return nil, nil
	}
	metrics := make([]*models.Metric, 0, len(metricsOld))
	for _, old := range metricsOld {
		m, err := old.ToNew()
		if err != nil {
			return nil, err
		}
		if m.Step == 0 {
			m.Step = int(p.Cycle)
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (p *Plugin) writeLog(logData []byte) {
	if p.LogFile == "" {
		return
	}
	os.MkdirAll(filepath.Dir(p.LogFile), os.ModePerm)
	f, err := os.OpenFile(p.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}
	f.WriteString(time.Now().Format(time.RFC3339Nano))
	f.WriteString("\t")
	f.Write(logData)
	f.Close()
}

// Hash return file bytes hash
func (p *Plugin) Hash() string {
	data, _ := ioutil.ReadFile(p.File)
	if len(data) > 0 {
		return utils.MD5HashBytes(data)
	}
	return ""
}

package logutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log = zaplog.Zap("logutil")

	writeFileHandler  *os.File
	writeFileRotate   bool
	writingFilename   string
	writingFileLayout string
)

// SetWriteFile sets writing filename
func SetWriteFile(file string) {
	if file == "" {
		return
	}
	if writeFileHandler != nil {
		writeFileHandler.Sync()
		writeFileHandler.Close()
	}
	if strings.Contains(file, "%s") {
		writingFileLayout = file
		writeFileRotate = true
		writingFilename = fmt.Sprintf(writingFileLayout, time.Now().Format("20060102"))
	} else {
		writingFilename = file
	}
	f, err := os.OpenFile(writingFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Warn("init-write-error", "error", err, "file", file)
		return
	}
	writeFileHandler = f
	log.Info("init-write", "file", writingFilename, "rotate", writeFileRotate)
}

// Stop stops metrics writing
func Stop() {
	if writeFileHandler != nil {
		writeFileHandler.Sync()
		writeFileHandler.Close()
		log.Info("stop")
	}
}

// Write writes metrics to file
func Write(metrics []*models.Metric) {
	if writeFileHandler == nil {
		return
	}
	buf := bytes.NewBuffer(nil)
	for _, m := range metrics {
		b, err := json.Marshal(m)
		if err != nil {
			continue
		}
		buf.Write(b)
		buf.WriteString("\n")
	}
	if buf.Len() == 0 {
		return
	}
	if writeFileHandler != nil {
		writeFileHandler.Write(buf.Bytes())
	}
	if rand.Intn(1000)%20 == 0 {
		tryRotate()
	}
}

func tryRotate() {
	newFile := fmt.Sprintf(writingFileLayout, time.Now().Format("20060102"))
	if newFile != writingFilename {
		log.Debug("do-rotate", "old", writingFilename, "new", newFile)
		// reset again
		SetWriteFile(writingFileLayout)
		go CleanOldRotated()
	}
}

var (
	// LogCleanDays means the log that over this days are cleaned
	LogCleanDays = 4
	// LogGzipDays means the log that over this days are gzipped
	LogGzipDays = 2
)

// CleanOldRotated cleans old rotated files
func CleanOldRotated() {
	for i := LogCleanDays + 2; i >= 1; i-- {
		tStr := time.Now().Add(time.Second * time.Duration(-86400*i)).Format("20060102")
		filename := fmt.Sprintf(writingFileLayout, tStr)
		if _, err := os.Stat(filename); err != nil {
			continue
		}
		if i >= LogCleanDays {
			os.Remove(filename)
			log.Info("do-remove", "file", filename)
			continue
		}
		if i >= LogGzipDays {
			gzFile := filename + ".gz"
			exec.Command("tar", "-czf", gzFile, filename).Run()
			os.Remove(filename)
			log.Info("do-gzip", "file", filename)
			continue
		}
	}
}

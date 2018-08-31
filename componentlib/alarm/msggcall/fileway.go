package msggcall

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
)

var (
	msggFileWay       string
	msggFilewayExpire int64 = 3600
	reqsFileLayout    string
	workDir, _        = os.Getwd()
)

var (
	filewayCallCount      = expvar.NewDiff("alert.fileway_call")
	filewayCallErrorCount = expvar.NewDiff("alert.fileway_call_error")
	filewayFilesCount     = expvar.NewDiff("alert.fileway_files")
)

func init() {
	expvar.Register(filewayCallCount, filewayCallErrorCount, filewayFilesCount)
}

const (
	filewaySuffix = ".alarms.json"
)

// SetFileway sets message fileway options
func SetFileway(file, layout string, expire int64) {
	msggFileWay = file
	if _, err := os.Stat(msggFileWay); err != nil {
		log.Warn("stat-msgg-fileway-error", "error", err)
		msggFileWay = ""
		return
	}
	os.MkdirAll(filepath.Dir(layout), 0755)
	reqsFileLayout = layout + filewaySuffix
	if expire > 0 {
		msggFilewayExpire = expire
	}
}

// CallFileWay writes requests to file and try to call script to handle the file
func CallFileWay(reqs []*msggRequest) {
	if reqsFileLayout == "" || msggFileWay == "" {
		return
	}
	file := fmt.Sprintf(reqsFileLayout, time.Now().Format("20060102_150405"))
	b, err := json.Marshal(reqs)
	if err != nil {
		log.Warn("callfile-json-error", "error", err)
		return
	}
	if err = ioutil.WriteFile(file, b, 0644); err != nil {
		log.Warn("callfile-write-error", "file", file, "error", err)
		return
	}
	log.Info("callfile-write", "file", file, "reqs", len(reqs))
	if msggFileWay != "" {
		absFile := filepath.Join(workDir, file)
		output, err := runWithTimeout(msggFileWay, []string{absFile}, time.Second*30)
		if err != nil {
			log.Warn("callfile-error", "file", absFile, "error", err)
			filewayCallErrorCount.Incr(1)
		} else {
			log.Info("callfile-ok", "file", absFile, "output", string(output))
			filewayCallCount.Incr(1)
		}
	}
	if rand.Intn(100)%30 == 0 { // use random ratio
		go cleanCallFile()
	}
}

func cleanCallFile() {
	var counts, deletes int64
	dir := filepath.Dir(reqsFileLayout)
	now := time.Now().Unix()
	filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(fpath, filewaySuffix) {
			return nil
		}
		if now-info.ModTime().Unix() > msggFilewayExpire {
			log.Info("callfile-remove", "file", fpath)
			os.RemoveAll(fpath)
			deletes++
		} else {
			counts++
		}
		return nil
	})
	filewayFilesCount.Set(counts)
	log.Info("callfile-removes", "file", counts, "deletes", deletes)
}

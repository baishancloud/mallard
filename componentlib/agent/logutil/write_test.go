package logutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCleanLog(t *testing.T) {
	Convey("clean-log", t, func() {
		writingFileLayout = "./tests/metrics_test_%s.json"
		prepareLogs(writingFileLayout)
		CleanOldRotated()
		for i := 0; i < LogCleanDays+2; i++ {
			tStr := time.Now().Add(time.Second * time.Duration(-86400*i)).Format("20060102")
			filename := fmt.Sprintf(writingFileLayout, tStr)
			if i >= LogGzipDays {
				filename = filename + ".gz"
			}
			_, err := os.Stat(filename)
			if i == LogCleanDays+1 {
				So(err, ShouldNotBeNil)
				continue
			}
			So(err, ShouldBeNil)
			os.Remove(filename)
		}
	})
	Convey("set.write", t, func() {
		SetWriteFile("./tests/test_%s.log", 4, 2)
		So(writingFilename, ShouldContainSubstring, time.Now().Format("20060102"))
		So(writeFileHandler, ShouldNotBeNil)

		Write([]*models.Metric{{
			Name:     "cpu",
			Time:     1,
			Value:    2,
			Endpoint: "localhost",
		}, {
			Name:     "cpu",
			Time:     2,
			Value:    2,
			Endpoint: "localhost",
		}})
		info, _ := os.Stat(writingFilename)
		So(info.Size(), ShouldEqual, 114)
		os.RemoveAll(writingFilename)

		SetWriteFile("./tests/ttt.log", 0, 0)
		So(writingFilename, ShouldEqual, "./tests/ttt.log")
		os.RemoveAll(writingFilename)

		Convey("stop.write", func() {
			Stop()
			So(writeFileHandler, ShouldBeNil)
		})

	})
}

func prepareLogs(layout string) {
	for i := 0; i < 8; i++ {
		tStr := time.Now().Add(time.Second * time.Duration(-86400*i)).Format("20060102")
		filename := fmt.Sprintf(writingFileLayout, tStr)
		ioutil.WriteFile(filename, make([]byte, 1024), 0644)
	}
}

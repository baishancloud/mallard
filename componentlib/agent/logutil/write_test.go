package logutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

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
}

func prepareLogs(layout string) {
	for i := 0; i < 8; i++ {
		tStr := time.Now().Add(time.Second * time.Duration(-86400*i)).Format("20060102")
		filename := fmt.Sprintf(writingFileLayout, tStr)
		ioutil.WriteFile(filename, make([]byte, 1024), 0644)
	}
}

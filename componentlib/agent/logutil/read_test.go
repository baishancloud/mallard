package logutil

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRead(t *testing.T) {
	Convey("read", t, func() {
		now := time.Now()
		Convey("read.file", func() {
			metrics, err := readFile("tests/test.log")
			So(err, ShouldBeNil)
			So(metrics, ShouldHaveLength, 1)
		})
		Convey("read.dir", func() {
			os.Chtimes("tests/test.log", now, now)
			SetReadDir("tests")
			metrics, err := readDirMetrics()
			So(err, ShouldBeNil)
			So(metrics, ShouldHaveLength, 1)
		})
		Convey("read.readed", func() {
			metrics, err := readFile("tests/test.log")
			So(err, ShouldBeNil)
			So(metrics, ShouldHaveLength, 0)
			So(files["tests/test.log"]-now.Unix()*1e9 < 1e9, ShouldBeTrue)
		})
	})
}

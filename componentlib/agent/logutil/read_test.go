package logutil

import (
	"os"
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
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
		Convey("read.chan", func() {
			os.Chtimes("tests/test.log", now, now.Add(time.Hour))
			ch := make(chan []*models.Metric, 100)
			readOnce(ch)()
			So(ch, ShouldHaveLength, 1)
		})
	})
}

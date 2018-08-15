package expvar

import (
	"os"
	"testing"

	"github.com/baishancloud/mallard/corelib/zaplog"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExpvars(t *testing.T) {
	Convey("expvars", t, func() {
		base := NewBase("base")
		base.Set(999)
		diff := NewDiff("diff")
		diff.Set(999)
		diff.Incr(99)
		qps := NewQPS("qps")
		avg := NewAverage("avg", 10)
		avg.Set(10)
		avg.Set(22)
		Register(base, diff, qps, avg)

		Convey("expose-factory", func() {
			values := ExposeFactory([]interface{}{base, diff, qps, avg}, false)
			So(values["base.cnt"], ShouldEqual, 999)
			So(values, ShouldContainKey, "diff.diff")
			So(values, ShouldContainKey, "qps.qps")
			So(values, ShouldContainKey, "avg.avg")
		})

		values := Expose(true)
		So(values, ShouldHaveLength, 4+9)
		So(values, ShouldContainKey, "procs.goroutine")

		Convey("print", func() {
			log = zaplog.Null()
			file := "test.log"
			printFunc(file, "metric", 120)()
			info, err := os.Stat(file)
			So(err, ShouldBeNil)
			So(info.Size() > 200, ShouldBeTrue)
			os.RemoveAll(file)
		})
	})
}

package expvar

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCounters(t *testing.T) {
	Convey("counters", t, func() {
		Convey("base", func() {
			b := NewBase("base")
			b.Set(123)
			So(b.Count(), ShouldEqual, 123)
			b.Incr(11)
			So(b.Count(), ShouldEqual, 134)
		})
		Convey("diff", func() {
			d := NewDiff("diff")
			d.Set(123)
			So(d.Diff(), ShouldEqual, 0)

			d.Set(128)
			So(d.Diff(), ShouldEqual, 5)
		})

		Convey("qps", func() {
			q := NewQPS("qps")
			q.Incr(120)
			So(q.QPS(), ShouldEqual, 0)

			q.Incr(120)
			q.lastTime = time.Now().Add(-1 * time.Minute)
			So(4-q.QPS() < 1e-5, ShouldBeTrue) // very 1 min diff
		})

		Convey("avg", func() {
			a := NewAverage("avg", 10)
			a.Set(123)
			So(a.Avg(), ShouldEqual, 123)
			a.Set(10)
			So(a.Avg(), ShouldEqual, float64(123*9+10)/10)
		})
	})
}

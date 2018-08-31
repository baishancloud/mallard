package syscollector

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDiskIO(t *testing.T) {
	Convey("disk.io", t, func() {

		metrics, err := DiskIOMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldBeGreaterThan, 0)

		So(metrics[0].Name, ShouldEqual, diskIOMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 11)
		So(metrics[0].Tags, ShouldContainKey, "device")
		So(metrics[0].Tags, ShouldContainKey, "mount")

		fmt.Println(metrics)
	})

	Convey("iostat", t, func() {
		IOStatsMetrics()
		time.Sleep(time.Second * 3)

		metrics, err := IOStatsMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldBeGreaterThan, 1)

		So(metrics[0].Name, ShouldEqual, diskIOStatMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 13)
		So(metrics[0].Tags, ShouldContainKey, "device")
		So(metrics[0].Fields, ShouldContainKey, "util")

		last := metrics[len(metrics)-1]
		So(last.Name, ShouldEqual, diskIOMaxMetricName)
		So(last.Fields, ShouldHaveLength, 0)

		fmt.Println(metrics)
	})
}

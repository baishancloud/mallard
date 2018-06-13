package syscollector

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDiskUsage(t *testing.T) {
	Convey("disk.usage", t, func() {

		metrics, err := DiskUsagesMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldBeGreaterThan, 1)

		So(metrics[0].Name, ShouldEqual, diskBytesMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 5)
		So(metrics[0].Tags, ShouldContainKey, "fstype")
		So(metrics[0].Tags, ShouldContainKey, "mount")

		So(metrics[1].Name, ShouldEqual, diskInodesMetricName)
		So(metrics[1].Fields, ShouldHaveLength, 5)
		So(metrics[1].Tags, ShouldContainKey, "fstype")
		So(metrics[1].Tags, ShouldContainKey, "mount")

		fmt.Println(metrics)
	})
}

package syscollector

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCPU(t *testing.T) {
	Convey("cpu", t, func() {
		CPUAllMetrics()
		time.Sleep(time.Second)

		metrics, err := CPUAllMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldBeGreaterThanOrEqualTo, runtime.NumCPU()+2)

		So(metrics[0].Name, ShouldEqual, cpuMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 12)

		So(metrics[1].Name, ShouldEqual, cpuCoreMetricName)
		So(metrics[1].Tags, ShouldContainKey, "core")
		So(metrics[1].Tags["core"], ShouldNotEqual, "all")
		So(metrics[1].Fields, ShouldHaveLength, 12)

		So(metrics[len(metrics)-1].Name, ShouldEqual, cpuCoreMockMetricName)
		fmt.Println(metrics)
	})
}

package syscollector

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMemory(t *testing.T) {
	Convey("mem", t, func() {

		metrics, err := MemoryMetrics()
		So(err, ShouldBeNil)
		So(metrics, ShouldHaveLength, 1)

		So(metrics[0].Name, ShouldEqual, meminfoMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 26)

		fmt.Println(metrics)
	})
}

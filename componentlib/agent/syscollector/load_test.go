package syscollector

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoad(t *testing.T) {
	Convey("load", t, func() {

		metrics, err := LoadMetrics()
		So(err, ShouldBeNil)
		So(metrics, ShouldHaveLength, 2)

		So(metrics[0].Name, ShouldEqual, loadAvgMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 5)

		So(metrics[1].Name, ShouldEqual, loadMiscMetricName)
		So(metrics[1].Fields, ShouldHaveLength, 4)

		fmt.Println(metrics)
	})
}

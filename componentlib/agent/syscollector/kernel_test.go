package syscollector

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKernel(t *testing.T) {
	Convey("kernel", t, func() {

		metrics, err := KernelMetrics()
		So(err, ShouldBeNil)
		So(metrics, ShouldHaveLength, 2)

		So(metrics[0].Name, ShouldEqual, kernelMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 3)

		So(metrics[1].Name, ShouldEqual, fileDescriptionMetricName)

		fmt.Println(metrics)
	})
}

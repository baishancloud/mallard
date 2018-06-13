package syscollector

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetSstat(t *testing.T) {
	Convey("netstat", t, func() {

		metrics, err := NetStatTCPExMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldEqual, 1)

		So(metrics[0].Name, ShouldEqual, netStatTCPExMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 117)

		fmt.Println(metrics)
	})
}

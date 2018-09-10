package syscollector

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetIface(t *testing.T) {
	Convey("netiface", t, func() {

		metrics, err := NetIfaceMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldBeGreaterThan, 0)

		So(metrics[0].Name, ShouldEqual, netIfaceMetricName)
		So(len(metrics[0].Fields), ShouldBeGreaterThanOrEqualTo, 18)
		So(metrics[0].Tags, ShouldContainKey, "iface")

		fmt.Println(metrics)
	})
}

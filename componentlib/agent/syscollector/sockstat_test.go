package syscollector

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSockstat(t *testing.T) {
	Convey("sockstat", t, func() {

		metrics, err := SockstatMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldEqual, 1)

		So(metrics[0].Name, ShouldEqual, netSockstatMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 8)

	})
}

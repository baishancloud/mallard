package syscollector

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSNMP(t *testing.T) {
	Convey("snmp.udp", t, func() {

		metrics, err := SnmpUDPMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldEqual, 1)

		So(metrics[0].Name, ShouldEqual, udpMetricName)
		So(len(metrics[0].Fields), ShouldBeGreaterThanOrEqualTo, 7)

		fmt.Println(metrics)
	})

	Convey("snmp.tcp", t, func() {
		SnmpTCPMetrics()
		time.Sleep(time.Second)

		metrics, err := SnmpTCPMetrics()
		So(err, ShouldBeNil)
		So(len(metrics), ShouldEqual, 2)

		So(metrics[0].Name, ShouldEqual, tcpMetricName)
		So(metrics[0].Fields, ShouldHaveLength, 18)

		So(metrics[1].Name, ShouldEqual, tcpPluginMetricName)

		fmt.Println(metrics)
	})
}

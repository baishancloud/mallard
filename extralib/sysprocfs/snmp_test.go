package sysprocfs

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSNMP(t *testing.T) {
	Convey("snmp", t, func() {
		tcp, err := SnmpTCP()
		So(err, ShouldBeNil)
		fmt.Println(tcp)

		time.Sleep(time.Second)

		tcp, err = SnmpTCP()
		So(err, ShouldBeNil)
		fmt.Println(tcp)

		udp, err := SnmpUDP()
		So(err, ShouldBeNil)
		fmt.Println(udp)
	})
}

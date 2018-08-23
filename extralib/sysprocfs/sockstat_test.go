package sysprocfs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSockstat(t *testing.T) {
	Convey("sockstat", t, func() {
		stat, err := Sockstat()
		So(err, ShouldBeNil)
		So(stat, ShouldContainKey, "tcp_inuse")
		So(stat, ShouldContainKey, "udp_inuse")
		So(stat, ShouldContainKey, "sockets_used")
		fmt.Println(stat)

		stat, err = Sockstat6()
		So(err, ShouldBeNil)
		So(stat, ShouldContainKey, "tcp6_inuse")
		So(stat, ShouldContainKey, "udp6_inuse")
		fmt.Println(stat)

		// So(stat, ShouldContainKey, "sockets_used")
	})
}

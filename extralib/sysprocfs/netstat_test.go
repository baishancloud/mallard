package sysprocfs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetstat(t *testing.T) {
	Convey("netstat", t, func() {
		conns, err := NetConnections()
		So(err, ShouldBeNil)
		fmt.Println(conns)

		nt, err := NetStat()
		So(err, ShouldBeNil)
		fmt.Println(nt)
	})
}

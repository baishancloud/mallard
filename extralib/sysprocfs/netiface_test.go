package sysprocfs

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetiface(t *testing.T) {
	Convey("netifaces", t, func() {
		netifaces, err := NetIfaceStats()
		So(err, ShouldBeNil)
		fmt.Println(netifaces)

		time.Sleep(time.Second * 2)

		netifaces, err = NetIfaceStats()
		So(err, ShouldBeNil)
		fmt.Println(netifaces)
	})
}

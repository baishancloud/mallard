package sysprocfs

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoad(t *testing.T) {
	Convey("load", t, func() {
		load, err := LoadAvg()
		So(err, ShouldBeNil)
		fmt.Println(load)
	})
	Convey("load_misc", t, func() {
		load, err := LoadMisc()
		So(err, ShouldBeNil)
		So(load.CtxtDiff, ShouldEqual, 0)
		fmt.Println(load)
		time.Sleep(time.Second)

		load, err = LoadMisc()
		So(err, ShouldBeNil)
		So(load.CtxtDiff, ShouldBeGreaterThanOrEqualTo, 0)
		fmt.Println(load)
	})
}

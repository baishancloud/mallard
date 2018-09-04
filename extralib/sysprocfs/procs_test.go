package sysprocfs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcsCount(t *testing.T) {
	Convey("procs", t, func() {
		count, err := ProcsCount()
		So(err, ShouldBeNil)
		So(count, ShouldBeGreaterThan, 0)
		fmt.Println(count)
	})
}

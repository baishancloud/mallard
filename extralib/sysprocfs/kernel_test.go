package sysprocfs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKernel(t *testing.T) {
	Convey("kernel", t, func() {
		kr, err := Kernel()
		So(err, ShouldBeNil)
		fmt.Println(kr)
	})
}

package sysprocfs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMemory(t *testing.T) {
	Convey("memory", t, func() {
		mem, err := Memory()
		So(err, ShouldBeNil)
		fmt.Println(mem)
	})
}

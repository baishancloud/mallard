package sysprocfs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSelfProcess(t *testing.T) {
	Convey("selfprocess", t, func() {
		stats, err := SelfProcessStat()
		So(err, ShouldBeNil)
		fmt.Println(stats)
	})
}

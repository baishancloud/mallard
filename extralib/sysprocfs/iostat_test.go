package sysprocfs

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIOstat(t *testing.T) {
	Convey("disk_io", t, func() {
		diskIO, err := DiskIO()
		So(err, ShouldBeNil)
		fmt.Println(diskIO)
	})
	Convey("io_stat", t, func() {
		ioStat, err := IOStats()
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 2)
		ioStat, err = IOStats()
		So(err, ShouldBeNil)
		fmt.Println(ioStat)
	})

}

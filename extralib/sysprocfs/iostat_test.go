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
		time.Sleep(time.Second * 5)
		ioStat, err = IOStats()
		So(err, ShouldBeNil)
		So(len(ioStat), ShouldBeGreaterThanOrEqualTo, 1)
		for _, ios := range ioStat {
			So(ios, ShouldNotBeNil)
		}
	})

}

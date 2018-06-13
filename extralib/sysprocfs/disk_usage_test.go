package sysprocfs

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDiskMount(t *testing.T) {
	Convey("disk_mount", t, func() {
		mt, err := DiskMounts()
		So(err, ShouldBeNil)
		fmt.Println(mt)
	})
}

func TestDiskUsage(t *testing.T) {
	Convey("disk_usage", t, func() {
		use, err := DiskUsages()
		So(err, ShouldBeNil)
		fmt.Println(use)
	})
}

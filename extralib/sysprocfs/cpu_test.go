package sysprocfs

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCPU(t *testing.T) {
	Convey("cpu_total", t, func() {
		_, err := CPUTotal()
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 2)
		value, err := CPUTotal()
		So(err, ShouldBeNil)
		So(value.CPU, ShouldEqual, cpuTotalName)
		fmt.Println(value)
	})
	Convey("cpu_core", t, func() {
		_, err := CPUCores()
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 2)
		values, err := CPUCores()
		So(err, ShouldBeNil)
		So(values, ShouldHaveLength, runtime.NumCPU())
		fmt.Println(values)
	})

}

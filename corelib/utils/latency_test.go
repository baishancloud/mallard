package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLatency(t *testing.T) {
	Convey("latency", t, func() {
		lat := NewLatency(3, 10)
		lat.Set(0, 100)
		lat.Set(1, 105)
		lat.Set(2, 99)
		lat.Set(3, 88)

		So(lat.History(), ShouldResemble, []int64{100, 105, 99})

		So(lat.Get(), ShouldEqual, 2)

		lat.SetFail(2)
		So(lat.Get(), ShouldEqual, 0)
		So(lat.Rand(), ShouldNotEqual, 2)

		v, err := lat.GetValue(1)
		So(err, ShouldBeNil)
		So(v, ShouldEqual, 105)

		lat.SetFail(0)
		So(lat.Get(), ShouldEqual, 1)

		lat.SetFail(1)
		lat.Get()

		v, err = lat.GetValue(2)
		So(err, ShouldBeNil)
		So(v, ShouldEqual, InitValue)

		_, err = lat.GetValue(99)
		So(err, ShouldEqual, ErrorOversize)
	})
}

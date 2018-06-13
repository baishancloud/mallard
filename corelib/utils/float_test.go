package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToFloat(t *testing.T) {
	Convey("tofloat64", t, func() {
		_, err := ToFloat64(nil)
		So(err, ShouldNotBeNil)

		_, err = ToFloat64("abc")
		So(err, ShouldNotBeNil)

		_, err = ToFloat64([]byte("abc"))
		So(err, ShouldNotBeNil)

		value, err := ToFloat64("123.456")
		So(err, ShouldBeNil)
		So(value, ShouldEqual, 123.456)

		value, err = ToFloat64(float64(122.22))
		So(err, ShouldBeNil)
		So(value, ShouldEqual, 122.22)

		value, err = ToFloat64(float32(122.22))
		So(err, ShouldBeNil)
		So(value-122.22, ShouldBeLessThan, 1e-3)

		value, err = ToFloat64(int(122))
		So(err, ShouldBeNil)
		So(value, ShouldEqual, 122.0)
	})
}

func TestFixFloat(t *testing.T) {
	Convey("fixfloat", t, func() {
		v := FixFloat(1.2444444, 2)
		So(v, ShouldEqual, 1.24)

		v = FixFloat(1.277777, 2)
		So(v, ShouldEqual, 1.27)
	})
}

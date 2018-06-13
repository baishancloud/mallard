package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMD5(t *testing.T) {
	Convey("md5", t, func() {
		So(MD5HashString("123456"), ShouldEqual, "e10adc3949ba59abbe56e057f20f883e")
	})
}

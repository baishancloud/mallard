package etcdapi

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient(t *testing.T) {
	Convey("client", t, func() {
		err := SetClient([]string{"http://localhost:2379"}, "", "", time.Second)
		So(err, ShouldBeNil)

		err = Set(context.Background(), "/abc/xyz", "abc.xyz")
		So(err, ShouldBeNil)

		err = Set(context.Background(), "/abc/abc", "abc.abc")
		So(err, ShouldBeNil)

		Convey("get", func() {
			value, err := Get(context.Background(), "/abc/xyz")
			So(value, ShouldEqual, "abc.xyz")
			So(err, ShouldBeNil)

			values, err := GetChildren(context.Background(), "/abc")
			So(values, ShouldHaveLength, 2)
			So(values["/abc/xyz"], ShouldEqual, "abc.xyz")
			So(values["/abc/abc"], ShouldEqual, "abc.abc")
			So(err, ShouldBeNil)
		}) 

		Convey("del", func() {
			err := Del(context.Background(), "/abc")
			So(err, ShouldBeNil)
		})
	})
}

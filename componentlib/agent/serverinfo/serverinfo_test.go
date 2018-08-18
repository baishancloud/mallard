package serverinfo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServerInfo(t *testing.T) {
	cachegroupFile = "test.conf"
	storageGroupFile = "test.conf"
	sertypesFile = "test.conf"
	hostnameFile = "test.conf"
	hostIDFile = "test.conf"
	Convey("read", t, func() {
		Read("default-ep", false)
		So(Cached(), ShouldNotBeNil)
		So(Hostname(), ShouldEqual, svrData.HostnameFirst)
		So(Sertypes(), ShouldEqual, "abcdfeghi")
		So(Cachegroup(), ShouldEqual, "abcdfeghi")
		So(StorageGroup(), ShouldEqual, "abcdfeghi")
		So(IP(), ShouldNotBeEmpty)

		svrData.Sertypes = ""
		SetSertypes("test-sertypes")
		So(Sertypes(), ShouldEqual, "test-sertypes")
	})
	Convey("sertypes", t, func() {
		Read("", false)
		So(Cached(), ShouldNotBeNil)
		So(Hostname(), ShouldEqual, svrData.HostnameOS)
		Read("", true)
		So(Cached(), ShouldNotBeNil)
		So(Hostname(), ShouldEqual, svrData.HostnameAllConf)
	})
}

package plugins

import (
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPlugins(t *testing.T) {
	Convey("plugins", t, func() {
		SetDir("./", "", []string{"tests"})
		ScanFiles()

		So(plugins, ShouldHaveLength, 1)
		So(plugins, ShouldContainKey, "tests/10_test.sh")

		queue := make(chan []*models.Metric, 1e3)
		execFiles(queue, time.Now().Unix()+100)
		time.Sleep(time.Second)
		So(queue, ShouldHaveLength, 1)

		So(Version(), ShouldBeBlank)

		hashes := FilesHash()
		So(hashes, ShouldHaveLength, 1)
		So(hashes, ShouldContainKey, "tests/10_test.sh")
	})
}

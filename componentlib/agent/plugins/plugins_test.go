package plugins

import (
	"os"
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPlugins(t *testing.T) {
	Convey("plugins", t, func() {
		SetDir("./", "", nil)
		So(pluginRunDirList, ShouldResemble, []string{"sys"})
		SetDir("./", "", []string{"tests", "test2"})
		ScanFiles()

		So(plugins, ShouldHaveLength, 3)
		So(plugins, ShouldContainKey, "tests/10_test.sh")
		So(plugins, ShouldContainKey, "tests/10_test_new.sh")
		So(plugins, ShouldContainKey, "tests/10_test_empty.sh")

		queue := make(chan []*models.Metric, 1e3)
		execFiles(queue, time.Now().Unix()+100)
		time.Sleep(time.Second)
		So(queue, ShouldHaveLength, 2)

		So(Version(), ShouldBeBlank)

		hashes := FilesHash()
		So(hashes, ShouldHaveLength, 3)
		So(hashes, ShouldContainKey, "tests/10_test.sh")

		os.RemoveAll("tests_10_test.sh.log")
		os.RemoveAll("tests_10_test_empty.sh.log")
	})
}

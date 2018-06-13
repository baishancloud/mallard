package plugins

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type fileInfo struct{}

func (m fileInfo) Name() string {
	return ""
}

func (m fileInfo) Size() int64 {
	return 0
}

func (m fileInfo) Mode() os.FileMode {
	return os.FileMode(0)
}

func (m fileInfo) ModTime() time.Time {
	return time.Time{}
}

func (m fileInfo) IsDir() bool {
	return false
}

func (m fileInfo) Sys() interface{} {
	return nil
}

func TestPlugin(t *testing.T) {
	Convey("plugin", t, func() {
		p, err := NewPlugin("300_test.py", "test.log", new(fileInfo))
		So(err, ShouldBeNil)

		So(time.Now().Unix()-p.LastExecTime, ShouldBeLessThan, 61)
		So(p.Cycle, ShouldEqual, 300)
		So(p.timeout, ShouldEqual, 299*time.Second)
		So(p.ShouldExec(time.Now().Unix()+300), ShouldBeTrue)

		Convey("plugin.exec", func() {
			p, err := NewPlugin("tests/10_test.sh", "", new(fileInfo))
			So(err, ShouldBeNil)
			metrics, err := p.Exec(parsedAsNew)
			So(err, ShouldBeNil)
			So(metrics, ShouldHaveLength, 1)
			So(p.Hash(), ShouldNotBeBlank)
		})

		Convey("plugin.nilfileinfo", func() {
			_, err := NewPlugin("tests/10_test.sh", "", nil)
			So(err, ShouldEqual, ErrNoFileInfo)

			_, err = NewPlugin("test.sh", "", nil)
			So(err, ShouldEqual, ErrWrongFilename)
		})
	})
}

package syscollector

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCollector(t *testing.T) {
	Convey("collector", t, func() {
		So(collectorFactory, ShouldHaveLength, 15)
	})
}

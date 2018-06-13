package processor

import (
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProccessor(t *testing.T) {
	Convey("processor", t, func(c C) {
		mQueue := make(chan []*models.Metric, 1e2)
		evtQueue := make(chan []*models.Event, 1e2)
		eQueue := make(chan error, 1e2)

		Register(func(metrics []*models.Metric) {
			c.So(metrics, ShouldHaveLength, 1)
		})
		RegisterEvent(func(evts []*models.Event) {
			c.So(evts, ShouldHaveLength, 1)
		})
		Process(mQueue, evtQueue, eQueue)
		mQueue <- []*models.Metric{{
			Name:     "abc",
			Value:    1,
			Endpoint: "localhost",
		}}
		mQueue <- []*models.Metric{{
			Name:     "abc",
			Value:    2,
			Endpoint: "localhost",
		}}
		mQueue <- []*models.Metric{{
			Name:     "abc",
			Value:    3,
			Endpoint: "localhost",
		}, nil}
		mQueue <- []*models.Metric{{
			Name:  "abc",
			Value: 4,
		}, new(models.Metric)}

		evtQueue <- []*models.Event{{}}
		time.Sleep(time.Second)
	})
}

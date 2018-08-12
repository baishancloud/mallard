package judger

import (
	"testing"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	log = zaplog.Null()
}

func TestJudger(t *testing.T) {
	Convey("judge", t, func() {
		ss := []*models.Strategy{
			{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#2)",
				Operator:       ">=",
				RightValue:     1,
			},
			{
				ID:             2,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#2)",
				Operator:       "<=",
				RightValue:     2,
			},
			{
				ID:             3,
				Metric:         "cpu",
				FieldTransform: "select(abcd)",
				Func:           "all(#2)",
				Operator:       "<=",
				RightValue:     2,
			},
			{
				ID:             4,
				Metric:         "cpu",
				FieldTransform: "select(abcd)",
				Func:           "all(#2)",
				Operator:       "<=",
				RightValue:     2,
				TagString:      "a=b",
			},
			{
				ID:             9,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "xyz(#2)",
				Operator:       "<=",
				RightValue:     2,
			},
		}
		events := SetStrategyData(ss)
		So(events, ShouldHaveLength, 0)
		So(unitsAccept["cpu"], ShouldHaveLength, 4)

		metrics := []*models.Metric{
			{
				Name:     "cpu",
				Time:     1,
				Value:    2,
				Endpoint: "localhost",
			},
			{
				Name:     "cpu",
				Time:     2,
				Value:    2,
				Endpoint: "localhost",
			},
			{
				Name:     "cpu",
				Time:     3,
				Value:    3,
				Endpoint: "localhost",
				Fields: map[string]interface{}{
					"abc": 0.1,
				},
			},
			{
				Name:     "cpu",
				Time:     4,
				Value:    3,
				Endpoint: "localhost",
				Fields: map[string]interface{}{
					"abc": 0.1,
				},
			},
			{
				Name:     "cpu",
				Time:     5,
				Value:    3,
				Endpoint: "localhost",
				Fields: map[string]interface{}{
					"abc": 0.1,
				},
			},
		}

		events = Judge(metrics)
		So(events, ShouldHaveLength, 8)
		So(events[0].Status, ShouldEqual, models.EventProblem)
		So(events[2].Fields, ShouldContainKey, "value")
		So(events[1].Status, ShouldEqual, models.EventProblem)
		So(events[2].Status, ShouldEqual, models.EventProblem)
		So(events[3].Status, ShouldEqual, models.EventOk)
		So(eventCurrent.All(), ShouldHaveLength, 2)

		events = Judge([]*models.Metric{
			{
				Name:     "cpu",
				Time:     4, // use old value, ignore it
				Value:    3,
				Endpoint: "localhost",
				Fields: map[string]interface{}{
					"abc": 0.1,
				},
			},
			{
				Name:     "cpu",
				Time:     6,
				Value:    3,
				Endpoint: "localhost",
				Fields: map[string]interface{}{
					"abc": 0.1,
				},
			},
			{
				Name:     "cpu",
				Time:     7,
				Value:    3,
				Endpoint: "localhost",
				Fields: map[string]interface{}{
					"abc": 0.1,
				},
			},
		})
		So(events, ShouldHaveLength, 4)
		for _, evt := range events {
			if evt.Status == models.EventOk && evt.Step > 3 {
				So(evt.Fields, ShouldHaveLength, 0) // event is simplified
			}
		}

		Convey("judge.close", func() {
			ss = ss[:1]
			events := SetStrategyData(ss)
			So(events, ShouldHaveLength, 1)
			So(events[0].Status, ShouldEqual, models.EventClosed)
		})
	})
}

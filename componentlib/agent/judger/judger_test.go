package judger

import (
	"testing"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

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
		}
		events := SetStrategyData(ss)
		So(events, ShouldHaveLength, 0)
		So(unitsAccept["cpu"], ShouldHaveLength, 2)

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
			},
		}

		events = Judge(metrics)
		So(events, ShouldHaveLength, 4)
		So(events[0].Status, ShouldEqual, models.EventProblem)
		So(events[1].Status, ShouldEqual, models.EventProblem)
		So(events[2].Status, ShouldEqual, models.EventProblem)
		So(events[3].Status, ShouldEqual, models.EventOk)

		Convey("judge.close", func() {
			ss = ss[:1]
			events := SetStrategyData(ss)
			So(events, ShouldHaveLength, 1)
		})
	})
}

package judger

import (
	"testing"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStrategyUnit(t *testing.T) {
	Convey("strategy.unit", t, func() {
		Convey("strategy.unit.invalid", func() {
			unit, err := NewUnit(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "xyz(#2)",
				Operator:       ">=",
				RightValue:     1,
			})
			So(err, ShouldNotBeNil)
			So(unit.ID(), ShouldEqual, 0)
			So(unit.GroupBy(), ShouldBeNil)

			unit, err = NewUnit(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#2)",
				Operator:       ">>",
				RightValue:     1,
			})
			So(err, ShouldNotBeNil)
			So(unit.ID(), ShouldEqual, 0)

			unit, err = NewUnit(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all#2", // all(#2 is valid, hehehe
				Operator:       "==",
				RightValue:     1,
			})
			So(err, ShouldNotBeNil)
			So(unit.ID(), ShouldEqual, 0)

			unit, err = NewUnit(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#x)",
				Operator:       "==",
				RightValue:     1,
			})
			So(err, ShouldNotBeNil)
			So(unit.ID(), ShouldEqual, 0)

			unit, err = NewUnit(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#2)",
				Operator:       "==",
				RightValue:     1,
				TagString:      "a=b,x",
			})
			So(err, ShouldNotBeNil)
			So(unit.ID(), ShouldEqual, 0)

			unit, err = NewUnit(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value(",
				Func:           "all(#2)",
				Operator:       "==",
				RightValue:     1,
			})
			So(err, ShouldNotBeNil)
			So(unit.ID(), ShouldEqual, 0)
		})

		Convey("strategy.unit.info", func() {
			unit, err := NewUnit(&models.Strategy{
				ID:             9,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#2)",
				Operator:       "==",
				RightValue:     1,
				TagString:      "a=b",
				Score:          0.5,
			})
			So(err, ShouldBeNil)
			So(unit.ID(), ShouldEqual, 9)
			So(unit.Operator(), ShouldNotBeNil)
			So(unit.Operator().Base(), ShouldNotBeNil)
			So(unit.Operator().Tags(), ShouldContainKey, "a")
			So(unit.GroupBy(), ShouldResemble, defaultUnitGroups)
			So(unit.Score(), ShouldEqual, 0.5)
		})

		unit, err := NewUnit(&models.Strategy{
			ID:             1,
			Metric:         "cpu",
			FieldTransform: "select(value)",
			Func:           "all(#2)",
			Operator:       ">=",
			RightValue:     1,
		})
		So(err, ShouldBeNil)
		So(unit.ID(), ShouldEqual, 1)

		Convey("strategy.unit.check", func() {
			leftValue, status, err := unit.Check(&models.Metric{
				Name:     "cpu",
				Time:     1,
				Value:    2,
				Endpoint: "localhost",
			}, "")
			So(err, ShouldBeNil)
			So(leftValue, ShouldEqual, 0)
			So(status, ShouldEqual, models.EventIgnore)

			leftValue, status, err = unit.Check(&models.Metric{
				Name:     "cpu",
				Time:     2,
				Value:    2,
				Endpoint: "localhost",
			}, "")
			So(err, ShouldBeNil)
			So(leftValue, ShouldEqual, 2)
			So(status, ShouldEqual, models.EventProblem)

			leftValue, status, err = unit.Check(&models.Metric{
				Name:     "cpu",
				Time:     3,
				Value:    -1,
				Endpoint: "localhost",
			}, "")
			So(err, ShouldBeNil)
			So(leftValue, ShouldEqual, -1)
			So(status, ShouldEqual, models.EventOk)
		})

		Convey("strategy.unit.accept", func() {
			err := unit.SetStrategy(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#2)",
				Operator:       ">=",
				RightValue:     1,
				TagString:      "core=all",
			})
			So(err, ShouldBeNil)

			So(unit.Accept(&models.Metric{
				Name:     "cpu",
				Time:     2,
				Value:    2,
				Endpoint: "localhost",
			}), ShouldBeFalse)
			So(unit.Accept(&models.Metric{
				Name:     "cpu",
				Time:     2,
				Value:    2,
				Endpoint: "localhost",
				Tags: map[string]string{
					"core": "all",
				},
			}), ShouldBeTrue)
			So(unit.Accept(&models.Metric{
				Name:     "load",
				Time:     2,
				Value:    2,
				Endpoint: "localhost",
				Tags: map[string]string{
					"core": "all",
				},
			}), ShouldBeFalse)

		})

		Convey("strategy.unit.tranform", func() {
			err := unit.SetStrategy(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(abc)",
				Func:           "all(#2)",
				Operator:       ">=",
				RightValue:     1,
			})
			So(err, ShouldBeNil)

			leftValue, status, err := unit.Check(&models.Metric{
				Name:     "cpu",
				Time:     4,
				Value:    2,
				Endpoint: "localhost",
			}, "")
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, FieldMissingError{})
			So(leftValue, ShouldEqual, 0)
			So(status, ShouldEqual, models.EventIgnore)
		})

		Convey("strategy.unit.queue", func() {
			err := unit.SetStrategy(&models.Strategy{
				ID:             1,
				Metric:         "cpu",
				FieldTransform: "select(value)",
				Func:           "all(#2)",
				Operator:       ">=",
				RightValue:     1,
			})
			So(err, ShouldBeNil)

			// add first data
			_, status, err := unit.Check(&models.Metric{
				Name:     "cpu",
				Time:     4,
				Value:    2,
				Endpoint: "new-localhost",
			}, "hash-2")
			So(err, ShouldBeNil)
			So(status, ShouldEqual, models.EventIgnore)
			So(unit.dataQueue["hash-2"], ShouldHaveLength, 1)

			// add second data
			_, status, err = unit.Check(&models.Metric{
				Name:     "cpu",
				Time:     14,
				Value:    2,
				Endpoint: "new-localhost",
			}, "hash-2")
			So(err, ShouldBeNil)
			So(status, ShouldEqual, models.EventProblem)
			So(unit.dataQueue["hash-2"], ShouldHaveLength, 2)

			_, status, err = unit.Check(&models.Metric{
				Name:     "cpu",
				Time:     14 + QueueDataExpiry + 10,
				Value:    2,
				Endpoint: "new-localhost",
			}, "hash-2")
			So(err, ShouldBeNil)
			So(unit.dataQueue["hash-2"], ShouldHaveLength, 1)
			So(status, ShouldEqual, models.EventIgnore)
		})
	})
}

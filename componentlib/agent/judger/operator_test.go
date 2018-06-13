package judger

import (
	"testing"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testHistoryData = []*models.EventValue{
		{
			Value: 1,
			Time:  100,
		},
		{
			Value: 2,
			Time:  200,
		},
		{
			Value: 3,
			Time:  300,
		},
		{
			Value: 5,
			Time:  400,
		},
		{
			Value: 7,
			Time:  500,
		},
		{
			Value: 11,
			Time:  600,
		},
		{
			Value: 13,
			Time:  700,
		},
		{
			Value: 17,
			Time:  800,
		},
	}
	testMetric = &models.Metric{
		Name:  "testttt",
		Value: 9,
		Tags: map[string]string{
			"user":  "xyz",
			"cache": "cache-1",
			"isp":   "dx",
		},
		Fields: map[string]interface{}{
			"x": 123,
			"y": 4.56,
			"z": 78.9,
		},
		Endpoint: "127-0-0-1",
		Step:     60,
		Time:     1999,
	}
)

func genSelect(funcStr, opStr string, rightValue float64) Operator {
	st, err := NewSelectFromStrategy(&models.Strategy{
		FieldTransform: "select(value)",
		Func:           funcStr,
		Operator:       opStr,
		RightValue:     rightValue,
	})
	if err != nil {
		panic(err)
	}
	return st
}

type accepts struct {
	Limit      int
	IsOK       bool
	LeftValue  float64
	FieldValue interface{}
}

func TestParseStrategy(t *testing.T) {
	Convey("strategu.rule", t, func() {
		Convey("strategy.rule.select", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user=xyz,name=aaa",
			})
			So(op, ShouldNotBeNil)
			So(err, ShouldBeNil)
			tags := op.Tags()
			So(tags, ShouldHaveLength, 2)
			So(tags, ShouldContainKey, "user")
			So(tags, ShouldContainKey, "name")
		})
		Convey("strategy.rule.rangeselect", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "rangeselect(x,10,100,y,1)",
				Func:           "all(#5)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user=xyz,name=aaa",
			})
			So(op, ShouldNotBeNil)
			So(err, ShouldBeNil)
			tags := op.Tags()
			So(tags, ShouldHaveLength, 2)
			So(tags, ShouldContainKey, "user")
			So(tags, ShouldContainKey, "name")

			_, err = FromStrategy(&models.Strategy{
				FieldTransform: "rangeselect(x,10,y,1)",
				Func:           "all(#5)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user=xyz,name=aaa",
			})
			So(err, ShouldNotBeNil)
		})

		_, err := FromStrategy(&models.Strategy{
			FieldTransform: "zzzzz(x)",
			Func:           "all(#3)",
			Operator:       "==",
			RightValue:     12,
			TagString:      "user=xyz,name=aaa",
		})
		So(err, ShouldNotBeNil)
	})
}

func TestCalculateFunc(t *testing.T) {
	Convey("calculate", t, func() {
		Convey("calculate.select-all", func() {
			// real values,1,2,3,5,7
			testCalOperator(genSelect("all(#5)", "<", 50), accepts{5, true, 1, 0})
			// real values,1,2,3,5,7,11
			testCalOperator(genSelect("all(#6)", "<", 3), accepts{6, false, 3, 0})
		})
		Convey("calculate.select-diff", func() {
			// real values, -1,-2,-4,-6,-10
			testCalOperator(genSelect("diff(#5)", "<", 1), accepts{6, true, -1, 0})
			// real values, -1,-2,-4,-6
			testCalOperator(genSelect("diff(#4)", "<", -5), accepts{5, false, -6, 0})
		})
		Convey("calculate.select-diffavg", func() {
			// real values, -1,-2,-4,-6,-10 average -23/5 = -4.6
			testCalOperator(genSelect("diffavg(#5)", "<", -3), accepts{6, true, -4.6, 0})
			// real values, -1,-2,-4,-6 average -13/4 = -3.25
			testCalOperator(genSelect("diffavg(#4)", "<", -5), accepts{5, false, -3.25, 0})
		})
		Convey("calculate.select-adiff", func() {
			// real values, 1,2,4,6,10
			testCalOperator(genSelect("adiff(#5)", ">", 0), accepts{6, true, 1, 0})
			// real values, 1,2,4,6
			testCalOperator(genSelect("adiff(#4)", ">", 5), accepts{5, false, 6, 0})
		})
		Convey("calculate.select-adiffavg", func() {
			// real values, 1,2,4,6,10 average 23/5 = 4.6
			testCalOperator(genSelect("adiffavg(#5)", "<", 5), accepts{6, true, 4.6, 0})
			// real values, 1,2,4,6 average 13/4 = 3.25
			testCalOperator(genSelect("adiffavg(#4)", ">=", 5), accepts{5, false, 3.25, 0})
		})
		Convey("calculate.select-rdiff", func() {
			// real values, 1,2,4,6,10
			testCalOperator(genSelect("rdiff(#5)", "<=", -1), accepts{6, true, -1, 0})
			// real values, 1,2,4,6
			testCalOperator(genSelect("rdiff(#4)", ">", 5), accepts{5, false, -1, 0})
		})
		Convey("calculate.select-rdiffavg", func() {
			// real values, 1,2,4,6,10 average 23/5 = 4.6
			testCalOperator(genSelect("rdiffavg(#5)", "<", 5), accepts{6, true, -2, 0})
			// real values, 1,2,4,6 average 13/4 = 3.25
			testCalOperator(genSelect("rdiffavg(#4)", ">=", 5), accepts{5, false, -1.5, 0})
		})
		Convey("calculate.select-pdiff", func() {
			// real values, 1,2,4,6,10
			testCalOperator(genSelect("pdiff(#6)", "<", 0), accepts{7, true, -100, 0})
			// real values, 1,2,4,6
			testCalOperator(genSelect("pdiff(#4)", ">", 5), accepts{5, false, -100, 0})
		})
		Convey("calculate.select-pdiffavg", func() {
			// real values, 1,2,4,6,10 average 23/5 = 4.6
			testCalOperator(genSelect("pdiffavg(#6)", ">", 5), accepts{7, false, -55.33189033189032, 0})
			// real values, 1,2,4,6 average 13/4 = 3.25
			testCalOperator(genSelect("pdiffavg(#4)", ">=", 5), accepts{5, false, -64.16666666666666, 0})
		})
		Convey("calculate.select-sum", func() {
			// real values, 1,2,3,5,7,11,13 sum 23
			testCalOperator(genSelect("sum(#7)", "=", 42), accepts{7, true, 42, 0})
			// real values, 1,2,3 sum 6
			testCalOperator(genSelect("sum(#3)", "<=", 5), accepts{3, false, 6, 0})
		})
		Convey("calculate.select-avg", func() {
			// real values, 1,2,3,5,7,11,13 average 42/7=6
			testCalOperator(genSelect("avg(#7)", "==", 6), accepts{7, true, 6, 0})
			// real values, 1,2,3 average  7/3=2.333
			testCalOperator(genSelect("avg(#3)", "<", 2), accepts{3, false, 7 / 3, 0})
		})
		Convey("calculate.select-have", func() {
			// real values, 1,2,3,5,7,11,13
			testCalOperator(genSelect("have(#7,6)", "!=", 1), accepts{7, true, 2, 0})
			// real values, 1,2,3
			testCalOperator(genSelect("have(#3,2)", "<", 2), accepts{3, false, 2, 0})
		})
	})
}

func TestOperator(t *testing.T) {
	Convey("operator", t, func() {
		Convey("operator.select", func() {
			// test select(value)
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(value)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
			})
			So(err, ShouldBeNil)
			testOperator(op, accepts{3, false, 1, testMetric.Value})

			op2, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user=xyz,name=aaa",
			})
			So(err, ShouldBeNil)
			testOperator(op2, accepts{3, false, 1, testMetric.Fields["x"]})
			So(op2.Tags(), ShouldContainKey, "user")
			So(op2.Tags(), ShouldContainKey, "name")
		})
		Convey("operator.rangeselect", func() {
			// test select(value)
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "rangeselect(value,0,10,value,-1)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
			})
			So(err, ShouldBeNil)
			testOperator(op, accepts{3, false, 1, testMetric.Value})

			op2, err := FromStrategy(&models.Strategy{
				FieldTransform: "rangeselect(value,0,10,x,-1)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
			})
			So(err, ShouldBeNil)
			testOperator(op2, accepts{3, false, 1, testMetric.Fields["x"]})

			op3, err := FromStrategy(&models.Strategy{
				FieldTransform: "rangeselect(x,10,99,x,-1)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
			})
			So(err, ShouldBeNil)
			testOperator(op3, accepts{3, false, 1, -1})
		})
	})
}

func testOperator(op Operator, acceptData accepts) {
	limit := op.Limit()
	So(limit, ShouldEqual, acceptData.Limit)
	fieldValue, err := op.Transform(testMetric)
	So(err, ShouldBeNil)
	So(fieldValue, ShouldEqual, acceptData.FieldValue)
	data := testHistoryData[:limit]
	leftValue, ok, err := op.Trigger(data)
	So(err, ShouldBeNil)
	So(ok, ShouldEqual, acceptData.IsOK)
	So(leftValue, ShouldEqual, acceptData.LeftValue)
}

func testCalOperator(operator Operator, acceptData accepts) {
	limit := operator.Limit()
	So(limit, ShouldEqual, acceptData.Limit)
	data := testHistoryData[:limit]
	leftValue, ok, err := operator.Trigger(data)
	So(err, ShouldBeNil)
	So(ok, ShouldEqual, acceptData.IsOK)
	So(leftValue, ShouldEqual, acceptData.LeftValue)
}

func accepetMetric(op Operator, metric *models.Metric) bool {
	tagFilters := op.Tags()
	if len(tagFilters) > 0 {
		fullTags := metric.FullTags()
		if len(fullTags) == 0 {
			return false
		}
		for key, tagFunc := range tagFilters {
			v := fullTags[key]
			if !tagFunc(v) {
				return false
			}
		}
	}
	return true
}

func TestTagFilter(t *testing.T) {
	Convey("tag.filter", t, func() {
		Convey("tag.filter.equal", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user=xyz,name=aaa",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xyz",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"name": "aaa",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xyz",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
		})
		Convey("tag.filter.not-equal", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user!=xyz,name=aaa",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xyz",
					"name": "aaa",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"name": "aaa",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"name": "aaa",
					"user": "xxx",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xyz",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
		})
		Convey("tag.filter.prefix", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user^=xy",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xyz",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xxx",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "zzzz",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
		})
		Convey("tag.filter.suffix", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "all(#3)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user$=z",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xyz",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xxx",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "zzzz",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
		})
		Convey("tag.filter.contains", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "have(#3,1)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user*=y",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "xyz",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "yyyxxxx",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "zzzzyyy",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "aaaaa",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
		})
		Convey("tag.filter.equal-some", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "have(#3,1)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user=[x|y|z]",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "x",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "y",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "z",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "aaaaa",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
		})
		Convey("tag.filter.not-equal-some", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "have(#3,1)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user!=[x|y|z]",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "x",
					"name": "aaa",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "y",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "z",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "aaaaa",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
		})
		Convey("tag.filter.prefix-some", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "have(#3,1)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user^=[x|y|z]",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "x",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "y",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "z",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "aaaaa",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
		})
		Convey("tag.filter.suffix-some", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "have(#3,1)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user$=[x|y|z]",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "x",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "ya",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "z",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "aaaaa",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
		})
		Convey("tag.filter.contains-some", func() {
			op, err := FromStrategy(&models.Strategy{
				FieldTransform: "select(x)",
				Func:           "have(#3,1)",
				Operator:       "==",
				RightValue:     12,
				TagString:      "user*=[x|aaaa|yz]",
			})
			So(err, ShouldBeNil)
			So(accepetMetric(op, testMetric), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "x",
					"name": "aaa",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "zayz0",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "z",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "aaaaa",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeTrue)
			So(accepetMetric(op, &models.Metric{
				Tags: map[string]string{
					"user": "aa",
					"name": "aaa",
					"uuu":  "xxx",
				},
			}), ShouldBeFalse)
		})
	})
}

package models

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var testM = &Metric{
	Name:     "load",
	Value:    1.23,
	Endpoint: "localhost",
	Time:     time.Now().Unix(),
	Fields: map[string]interface{}{
		"1min":  1.23,
		"5min":  2.22,
		"15min": 1.98,
	},
	Tags: map[string]string{
		"node": "node-1",
		"user": "monitor-host",
	},
}

var testMRaw = &MetricRaw{
	Metric:    "load",
	Value:     1.23,
	Endpoint:  "localhost",
	Timestamp: time.Now().Unix(),
	Fields:    "1min=1.2,5min=2.5,15min=0.8",
	Tags:      "node=node-1,user=monitor-host",
}

func TestEntityMetric(t *testing.T) {
	Convey("metric", t, func() {
		So(testM.TagString(false), ShouldEqual, "name=load,node=node-1,user=monitor-host")
		So(testM.TagString(true), ShouldEqual, "endpoint=localhost,name=load,node=node-1,user=monitor-host")
		So(testM.Hash(), ShouldEqual, "ec4d166aaadba9e7fd59d64b79a6626b421aa90e")

		testM.FillTags("mallard-node", "mallard-cache", "mallard-storage", "abc")
		So(testM.Hash(), ShouldEqual, "ec4d166aaadba9e7fd59d64b79a6626b421aa90e")

		fullTags := testM.FullTags()
		So(fullTags["sertypes"], ShouldEqual, "mallard-node")
		So(fullTags["endpoint"], ShouldEqual, testM.Endpoint)
		So(fullTags["agent_endpoint"], ShouldEqual, "abc")
	})

	Convey("metric.raw", t, func() {
		m, err := testMRaw.ToNew()
		So(err, ShouldBeNil)
		So(m.Fields, ShouldHaveLength, 3)
		So(m.Tags, ShouldHaveLength, 2)

		testMRaw.Fields += ",xvz"
		_, err = testMRaw.ToNew()
		So(err, ShouldNotBeNil)
	})
}

func BenchmarkEntityMetricTagString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testM.TagString(true)
	}
}

func BenchmarkEntityMetricHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testM.Hash()
	}
}

package judgestore

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/baishancloud/mallard/componentlib/judge/judgestore/filter"
	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

var nowUnix = time.Now().Unix()
var metrics = []*models.Metric{
	{
		Name:     "cpu",
		Time:     nowUnix,
		Value:    2,
		Endpoint: "localhost",
	},
	{
		Name:     "cpu",
		Time:     nowUnix,
		Value:    2,
		Endpoint: "localhost",
	},
	{
		Name:     "cpu",
		Time:     nowUnix,
		Value:    3,
		Endpoint: "localhost",
		Fields: map[string]interface{}{
			"abc": 0.1,
		},
	},
	{
		Name:     "cpu",
		Time:     nowUnix,
		Value:    3,
		Endpoint: "localhost",
		Fields: map[string]interface{}{
			"abc": 0.1,
		},
	},
	{
		Name:     "cpu",
		Time:     nowUnix,
		Value:    3,
		Endpoint: "localhost",
		Fields: map[string]interface{}{
			"abc": 0.1,
		},
	},
}

var expiredMetrics = []*models.Metric{
	{
		Name:     "memory",
		Time:     nowUnix,
		Value:    2,
		Endpoint: "localhost",
	},
	{
		Name:     "memory",
		Time:     nowUnix,
		Value:    2,
		Endpoint: "localhost",
		Tags:     map[string]string{"a": "a"},
	},
	{
		Name:     "memory",
		Time:     nowUnix - 86400,
		Value:    2,
		Endpoint: "localhost",
	},
}

func TestStore(t *testing.T) {
	Convey("store", t, func() {
		Convey("write-metrics", func() {
			SetDir("test-store")
			Write(metrics)
			shouldFile := fmt.Sprintf("test-store/%s.log", storeKey(metrics[0]))
			info, err := os.Stat(shouldFile)
			So(err, ShouldBeNil)
			So(info.Size() > 390, ShouldBeTrue)

			syncFileHandlers()

			os.RemoveAll("test-store")

			Close()
			Write(metrics)
			_, err = os.Stat(shouldFile)
			So(err, ShouldNotBeNil)
		})

		Convey("ressemble-metrics", func() {
			results := ressembleMetrics(metrics)
			So(results, ShouldHaveLength, 1)
			So(results[storeKey(metrics[0])], ShouldHaveLength, 5)

			results = ressembleMetrics(expiredMetrics)
			So(results, ShouldHaveLength, 1)
			So(results[storeKey(expiredMetrics[0])], ShouldHaveLength, 2)

			testFilters := make(filter.ForMetrics)
			testFilters["memory"] = &filter.ForMetric{
				Name:   "memory",
				Expire: 86400,
				Tags:   map[string]bool{"a": true},
			}
			SetFilters(testFilters)

			results = ressembleMetrics(expiredMetrics)
			So(results, ShouldHaveLength, 2)
			So(results[storeKey(expiredMetrics[0])], ShouldHaveLength, 2)
			So(results[storeKey(expiredMetrics[2])], ShouldHaveLength, 1)
		})

	})
}

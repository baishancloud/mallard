package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	go Listen("127.0.0.1:41234")
	time.Sleep(time.Second)
}

func TestGet(t *testing.T) {
	Convey("get-config", t, func() {
		resp, err := http.Get("http://127.0.0.1:41234/v1/config")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)
		err = json.NewDecoder(resp.Body).Decode(new(models.EndpointData))
		So(err, ShouldBeNil)
	})
	Convey("get-event", t, func() {
		resp, err := http.Get("http://127.0.0.1:41234/v1/event")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)
	})
	Convey("get-version", t, func() {
		resp, err := http.Get("http://127.0.0.1:41234/v1/version")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)
	})
}

func TestPushMetrics(t *testing.T) {
	Convey("push-metrics", t, func() {
		Convey("v1", func() {
			body := ``
			resp, err := http.Post("http://127.0.0.1:41234/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)

			body = `[{"metric": "avload","endpoint": "localhost", "value": 0,"fields": "avg_1min=0.58,avg_5min=0.78,avg_30min=0.77,cpu_core_count=24,serious=0,free_mem=34.3189","timestamp": 1531194364,"counterType": "GAUGE","step": 60}]`
			resp, err = http.Post("http://127.0.0.1:41234/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)

			body = `[{"metric": "avload","endpoint": "localhost", "value": 0,"fields": "avg_1min=0.58,avg_5min=0.78,avg_30min=0.77,cpu_core_count=24,serious=0,free_mem=34.3189","timestamp": 1531194364,"counterType": `
			resp, err = http.Post("http://127.0.0.1:41234/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)

			body = `[{"metric": "avload","endpoint": "localhost", "value": 0,"fields": "avg_1min=0.58,avg_5min0.78,avg_30min=0.77,cpu_core_count=24,serious=0,free_mem=34.3189","timestamp": 1531194364,"counterType": "GAUGE","step": 60}]`
			resp, err = http.Post("http://127.0.0.1:41234/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)
		})
		Convey("v2", func() {
			body := ``
			resp, err := http.Post("http://127.0.0.1:41234/v2/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)

			body = `[{"name":"mallard2_agent_perf","time":1531194441,"fields":{"event.strategy.cnt":1928,"events.ok.diff":3480,"events.problem.diff":0,"http.metric_fail.diff":0,"http.metric_recv.diff":0,"http.metric_req.diff":0,"logtool.read.diff":1,"plugins.cnt":0,"plugins.exec.diff":0,"plugins.fail.diff":0,"poster.config_change.diff":0,"poster.config_fail.diff":0,"poster.config_req.diff":4,"poster.event.diff":3480,"poster.event_fail.diff":0,"poster.event_latency.avg":0.27,"poster.metric.diff":4441,"poster.metric_fail.diff":0,"poster.metric_latency.avg":0.62,"procs.cpu":1.66,"procs.fds":8,"procs.goroutine":17,"procs.mem":21.64,"procs.mem_percent":1.08,"procs.read":10020,"procs.read_bytes":12288,"procs.write":873,"procs.write_bytes":0,"sys.collect.diff":4440},"step":120}]`
			resp, err = http.Post("http://127.0.0.1:41234/v2/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)

			body = `[{"name":"mallard2_agent_perf","time":1531194441,"fields":{"event.strategy.cnt":1928,"events.ok.diff":3480,"events.problem.diff":0,"http.metric_fail.diff":0,"http.`
			resp, err = http.Post("http://127.0.0.1:41234/v2/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)
		})
	})
}

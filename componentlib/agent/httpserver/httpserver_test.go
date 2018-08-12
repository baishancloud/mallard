package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	log = zaplog.Null()
}

var (
	server *httptest.Server
)

func setupServer() func() {
	server = httptest.NewServer(CreateHandlers())
	return func() {
		server.Close()
	}
}

func TestGet(t *testing.T) {
	defer setupServer()()
	Convey("get-config", t, func() {
		resp, err := http.Get(server.URL + "/v1/config")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)
		err = json.NewDecoder(resp.Body).Decode(new(models.EndpointData))
		So(err, ShouldBeNil)
	})
	Convey("get-event", t, func() {
		resp, err := http.Get(server.URL + "/v1/event")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)

		resp, err = http.Get(server.URL + "/v1/event?strategy=101")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)

		resp, err = http.Get(server.URL + "/v1/event?strategy=xyz")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 500)
	})
	Convey("get-version", t, func() {
		SetVersion("9.9.9")
		resp, err := http.Get(server.URL + "/v1/version")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, 200)
		result := make(map[string]string)
		err = json.NewDecoder(resp.Body).Decode(&result)
		So(err, ShouldBeNil)
		So(result["version"], ShouldEqual, "9.9.9")
		So(result, ShouldContainKey, "plugin")
	})
}

func TestPushMetrics(t *testing.T) {
	defer setupServer()()
	Convey("push-metrics", t, func() {
		Convey("v1", func() {
			q := make(chan []*models.Metric, 1e4)
			SetQueue(q)

			body := ``
			resp, err := http.Post(server.URL+"/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)

			body = `[{"metric": "avload","endpoint": "localhost", "value": 0,"fields": "avg_1min=0.58,avg_5min=0.78,avg_30min=0.77,cpu_core_count=24,serious=0,free_mem=34.3189","timestamp": 1531194364,"counterType": "GAUGE","step": 60}]`
			resp, err = http.Post(server.URL+"/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			So(len(q), ShouldEqual, 1)

			body = `[{"metric": "avload","endpoint": "localhost", "value": 0,"fields": "avg_1min=0.58,avg_5min=0.78,avg_30min=0.77,cpu_core_count=24,serious=0,free_mem=34.3189","timestamp": 1531194364,"counterType": `
			resp, err = http.Post(server.URL+"/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)
			So(len(q), ShouldEqual, 1)

			body = `[{"metric": "avload","endpoint": "localhost", "value": 0,"fields": "avg_1min=0.58,avg_5min0.78,avg_30min=0.77,cpu_core_count=24,serious=0,free_mem=34.3189","timestamp": 1531194364,"counterType": "GAUGE","step": 60}]`
			resp, err = http.Post(server.URL+"/v1/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)
			So(len(q), ShouldEqual, 1)
		})
		Convey("v2", func() {
			body := ``
			resp, err := http.Post(server.URL+"/v2/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)

			body = `[{"name":"mallard2_agent_perf","time":1531194441,"fields":{"event.strategy.cnt":1928,"events.ok.diff":3480,"events.problem.diff":0,"http.metric_fail.diff":0,"http.metric_recv.diff":0,"http.metric_req.diff":0,"logtool.read.diff":1,"plugins.cnt":0,"plugins.exec.diff":0,"plugins.fail.diff":0,"poster.config_change.diff":0,"poster.config_fail.diff":0,"poster.config_req.diff":4,"poster.event.diff":3480,"poster.event_fail.diff":0,"poster.event_latency.avg":0.27,"poster.metric.diff":4441,"poster.metric_fail.diff":0,"poster.metric_latency.avg":0.62,"procs.cpu":1.66,"procs.fds":8,"procs.goroutine":17,"procs.mem":21.64,"procs.mem_percent":1.08,"procs.read":10020,"procs.read_bytes":12288,"procs.write":873,"procs.write_bytes":0,"sys.collect.diff":4440},"step":120}]`
			resp, err = http.Post(server.URL+"/v2/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)

			body = `[{"name":"mallard2_agent_perf","time":1531194441,"fields":{"event.strategy.cnt":1928,"events.ok.diff":3480,"events.problem.diff":0,"http.metric_fail.diff":0,"http.`
			resp, err = http.Post(server.URL+"/v2/push", "application/json", strings.NewReader(body))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 500)
		})
	})
}

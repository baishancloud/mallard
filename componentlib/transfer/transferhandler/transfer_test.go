package transferhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	server *httptest.Server
	client = &http.Client{
		Timeout: time.Second * 5,
	}
	metricValues = make([]*models.Metric, 100)
	eventsValues = make([]*models.Event, 100)

	tMetricQueue = queues.NewQueue(1e6, "")
	tEvtQueue    = queues.NewQueue(1e6, "")
)

func createRequest(method, url string, data interface{}, withAuth bool) *http.Request {
	buf, _ := utils.GzipJSON(data, 1024)
	req, _ := http.NewRequest(method, url, buf)
	req.Header.Set("Agent-Endpoint", "hostname")
	req.Header.Set("Data-Length", "100")
	req.Header.Set("User-Agent", "mallard2-tester")
	if withAuth {
		for k, v := range httptoken.BuildHeader("test-token") {
			req.Header.Add(k, v)
		}
	}
	return req
}

func setupServer() func() {
	server = httptest.NewServer(Create(true, true))
	return func() {
		server.Close()
	}
}

func TestTransferHandler(t *testing.T) {
	SetQueues(tMetricQueue, tEvtQueue)
	Convey("transfer", t, func() {
		defer setupServer()()
		Convey("push-metrics", func() {
			req := createRequest("POST", server.URL+"/api/metric", metricValues, true)
			So(req, ShouldNotBeNil)
			resp, err := client.Do(req)
			So(resp.StatusCode, ShouldEqual, 204)
			So(err, ShouldBeNil)
			So(tMetricQueue.Len(), ShouldEqual, 1)
			So(metricsRecvQPS.Count(), ShouldEqual, 100)
		})
		Convey("push-metrics-401", func() {
			req := createRequest("POST", server.URL+"/api/metric", metricValues, false)
			So(req, ShouldNotBeNil)
			resp, err := client.Do(req)
			So(resp.StatusCode, ShouldEqual, 401)
			So(err, ShouldBeNil)
			So(tMetricQueue.Len(), ShouldEqual, 1)
		})
		Convey("push-events", func() {
			req := createRequest("POST", server.URL+"/api/event", eventsValues, true)
			So(req, ShouldNotBeNil)
			resp, err := client.Do(req)
			So(resp.StatusCode, ShouldEqual, 204)
			So(err, ShouldBeNil)
			So(tEvtQueue.Len(), ShouldEqual, 1)
			So(eventRecvQPS.Count(), ShouldEqual, 100)
		})
		Convey("push-no-datalength", func() {
			req := createRequest("POST", server.URL+"/api/event", eventsValues, true)
			req.Header.Del("Data-Length")
			So(req, ShouldNotBeNil)
			resp, err := client.Do(req)
			So(resp.StatusCode, ShouldEqual, 204)
			So(err, ShouldBeNil)
			So(tEvtQueue.Len(), ShouldEqual, 2)
			So(eventRecvQPS.Count(), ShouldEqual, 101)

			req = createRequest("POST", server.URL+"/api/metric", metricValues, true)
			req.Header.Del("Data-Length")
			So(req, ShouldNotBeNil)
			resp, err = client.Do(req)
			So(resp.StatusCode, ShouldEqual, 204)
			So(err, ShouldBeNil)
			So(tMetricQueue.Len(), ShouldEqual, 2)
			So(metricsRecvQPS.Count(), ShouldEqual, 101)
		})

		Convey("404", func() {
			req := createRequest("POST", server.URL+"/api/xxx", metricValues, true)
			resp, err := client.Do(req)
			So(resp.StatusCode, ShouldEqual, 404)
			So(err, ShouldBeNil)
		})

		Convey("get-config", func() {
			req := createRequest("GET", server.URL+"/api/config?ep=hostname&hash=abc", nil, true)
			resp, err := client.Do(req)
			So(resp.StatusCode, ShouldEqual, 404)
			So(err, ShouldBeNil)
			So(resp.Header.Get("Transfer-Time"), ShouldNotBeEmpty)

			req = createRequest("GET", server.URL+"/api/config?epxxxx=hostname&hash=abc", nil, true)
			resp, err = client.Do(req)
			So(resp.StatusCode, ShouldEqual, 500)
			So(err, ShouldBeNil)
		})
	})
}

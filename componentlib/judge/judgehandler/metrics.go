package judgehandler

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	m "github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/influxdata/influxdb/models"
	"github.com/julienschmidt/httprouter"
)

var (
	metricsReqQPS  = expvar.NewQPS("http.metrics_req")
	metricsRecvQPS = expvar.NewQPS("http.metrics_recv")
	//packReqQPS     = expvar.NewQPS("http.packs_req")
	//packRecvQPS    = expvar.NewQPS("http.packs_revc")
	influxReqQPS  = expvar.NewQPS("http.influx_req")
	influxRecvQPS = expvar.NewQPS("http.influx_recv")
)

func init() {
	expvar.Register(metricsRecvQPS, metricsReqQPS, influxRecvQPS, influxReqQPS)
}

func metricsRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	metricsReqQPS.Incr(1)
	dataLen, _ := strconv.ParseInt(r.Header.Get("Data-Length"), 10, 64)
	metrics := make([]*m.Metric, 0, dataLen)
	if err := httputil.LoadJSON(r, &metrics); err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	if len(metrics) == 0 {
		httputil.ResponseFail(rw, r, errors.New("no-metrics"))
		return
	}
	metricsRecvQPS.Incr(int64(len(metrics)))
	if queue != nil {
		queue <- metrics
	}
	rw.WriteHeader(204)
	log.Debug("recv-ok",
		"len", dataLen,
		"gzip", r.Header.Get("Content-Length"),
		"remote", r.RemoteAddr)
}

func influxRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	influxReqQPS.Incr(1)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httputil.ResponseErrorJSON(rw, r, 400, err)
		return
	}
	percision := r.URL.Query().Get("precision")
	if percision == "" {
		percision = "s"
	}
	points, err := models.ParsePointsWithPrecision(body, time.Now().UTC(), percision)
	if err != nil {
		httputil.ResponseErrorJSON(rw, r, 400, err)
		return
	}
	mlist := make([]*m.Metric, 0, len(points))
	for _, p := range points {
		m := &m.Metric{
			Name: string(p.Name()),
			Time: p.Time().Unix(),
			Tags: p.Tags().Map(),
		}
		m.Endpoint = m.Tags["endpoint"]
		delete(m.Tags, "endpoint")
		fields, err := p.Fields()
		if err != nil {
			log.Warn("fields-error", "metric", p.Name())
			continue
		}
		m.Fields = fields
		m.Value, _ = utils.ToFloat64(m.Fields["value"])
		delete(m.Fields, "value")

		if m.Name == "" || m.Time == 0 {
			log.Warn("fields-error-value", "metric", p.Name())
			continue
		}
		mlist = append(mlist, m)
	}
	if len(mlist) == 0 {
		httputil.ResponseErrorJSON(rw, r, 400, errors.New("empty-points-to-metric"))
		return
	}
	if len(mlist) != len(points) {
		log.Warn("influx-recv-convert-dismatch", "total", len(mlist), "points", len(points))
	}
	if queue != nil {
		queue <- mlist
	}
	rw.WriteHeader(204)
	dataLen := int64(len(mlist))
	influxRecvQPS.Incr(dataLen)
	/*log.Debug("influx-ok",
	"len", dataLen,
	"remote", r.RemoteAddr)*/
}

package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/agent/judger"
	"github.com/baishancloud/mallard/componentlib/agent/plugins"
	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/componentlib/agent/transfer"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/pprofwrap"
	"github.com/julienschmidt/httprouter"
)

var (
	mQueue chan<- []*models.Metric
)

func init() {
	expvar.Register(metricFailCount, metricRecvCount, metricReqCount)
}

// SetQueue sets metric channel for http recieving
func SetQueue(mCh chan<- []*models.Metric) {
	mQueue = mCh
}

func initHandlers() http.Handler {
	r := httprouter.New()
	r.GET("/v1/config", configGet)
	r.GET("/v1/event", eventGet)
	r.POST("/v1/push", metricsRecv)
	r.POST("/v2/push", metricsNewRecv)
	r.GET("/v1/version", versionGet)
	r.HandlerFunc("GET", "/debug/vars", expvar.HTTPHandler)
	r = pprofwrap.Wrap(r)
	return r
}

func responseFail(rw http.ResponseWriter, err error, msg string) {
	rw.WriteHeader(500)
	if err != nil {
		rw.Write([]byte(err.Error()))
		log.Warn(msg, "err", err.Error())
	}
}

func readBodyJSON(r *http.Request, value interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(value)
}

func writeBodyJSON(rw http.ResponseWriter, value interface{}) {
	encoder := json.NewEncoder(rw)
	encoder.Encode(value)
}

var (
	metricRecvCount = expvar.NewDiff("http.metric_recv")
	metricReqCount  = expvar.NewDiff("http.metric_req")
	metricFailCount = expvar.NewDiff("http.metric_fail")
)

func metricsRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	metricReqCount.Incr(1)
	var oldMs []*models.MetricRaw
	if err := readBodyJSON(r, &oldMs); err != nil {
		responseFail(rw, err, "recv error")
		metricFailCount.Incr(1)
		return
	}
	if len(oldMs) > 0 {
		ms := make([]*models.Metric, 0, len(oldMs))
		for _, m := range oldMs {
			m2, err := m.ToNew()
			if err != nil {
				responseFail(rw, err, "recv error")
				return
			}
			ms = append(ms, m2)
		}
		if mQueue != nil {
			mQueue <- ms
		}
		metricRecvCount.Incr(int64(len(ms)))
		record := make(map[string]struct{})
		for _, m := range ms {
			record[m.Name] = struct{}{}
		}
		log.Info("recv", "len", len(ms), "name", keysOfMap(record))
	}
	rw.WriteHeader(200)
	rw.Write([]byte("OK"))
}

func metricsNewRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	metricReqCount.Incr(1)
	var ms []*models.Metric
	if err := readBodyJSON(r, &ms); err != nil {
		responseFail(rw, err, "recv new error")
		metricFailCount.Incr(1)
		return
	}
	if len(ms) > 0 {
		if mQueue != nil {
			mQueue <- ms
		}
		metricRecvCount.Incr(int64(len(ms)))
		record := make(map[string]struct{})
		for _, m := range ms {
			record[m.Name] = struct{}{}
		}
		log.Info("recv", "len", len(ms), "name", keysOfMap(record))
	}
	rw.WriteHeader(200)
	rw.Write([]byte("OK"))
}

func configGet(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	encoder := json.NewEncoder(rw)
	encoder.Encode(transfer.EndpointData())
}

func eventGet(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	history := judger.Currents()
	if history == nil {
		rw.WriteHeader(404)
		return
	}
	if sid := r.FormValue("strategy"); sid != "" {
		id, _ := strconv.Atoi(sid)
		events := history.FindByStrategy(id)
		writeBodyJSON(rw, events)
		return
	}
	events := history.All()
	writeBodyJSON(rw, events)
}

var versionString string

// SetVersion sets version string to show in http response
func SetVersion(version string) {
	versionString = version
}

func versionGet(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	m := map[string]string{
		"version":  versionString,
		"ip":       serverinfo.IP(),
		"hostname": serverinfo.Hostname(),
		"plugin":   plugins.Version(),
	}
	encoder := json.NewEncoder(rw)
	encoder.Encode(m)
}

func keysOfMap(m map[string]struct{}) []string {
	if len(m) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

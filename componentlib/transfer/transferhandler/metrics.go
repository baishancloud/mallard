package transferhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/julienschmidt/httprouter"
)

var (
	metricsReqQPS      = expvar.NewQPS("http.metric_req")
	metricsRecvQPS     = expvar.NewQPS("http.metric_recv")
	metricsOpenReqQPS  = expvar.NewQPS("http.metric_open_req")
	metricsOpenRecvQPS = expvar.NewQPS("http.metric_open_recv")

	metricsRopQPS     = expvar.NewQPS("http.metric_pop")
	metricsRopZeroQPS = expvar.NewQPS("http.metric_pop_zero")
	metricsPopDataQPS = expvar.NewQPS("http.metric_pop_data")

	storePushQPS          = expvar.NewQPS("store.push")
	storePopAvg           = expvar.NewAverage("store.pop_size", 50)
	storePopFailDiff      = expvar.NewDiff("store.pop_fail")
	storeDropDiff         = expvar.NewDiff("store.drop")
	storeQueueLengthCount = expvar.NewBase("store.queue_length")
)

func init() {
	expvar.Register(metricsReqQPS, metricsRecvQPS, metricsOpenReqQPS, metricsOpenRecvQPS,
		metricsPopDataQPS, metricsRopQPS, metricsRopZeroQPS,
		storePushQPS, storePopFailDiff, storeDropDiff, storeQueueLengthCount, storePopAvg)
}

var mQueue *queues.Queue

var (
	// ErrMetricsPushFail means failure when pushing metrics to queue
	ErrMetricsPushFail = errors.New("metrics-push-fail")
)

func metricsRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	metricsReqQPS.Incr(1)
	pack, err := httputil.LoadPack(r, 1024*3)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	if mQueue != nil {
		dump, ok := mQueue.Push(*pack)
		if !ok {
			httputil.ResponseFail(rw, r, ErrMetricsPushFail)
			log.Warn("m-recv-error", "err", err, "remote", r.RemoteAddr)
			storeDropDiff.Incr(int64(pack.Len))
			return
		}
		if dump > 0 {
			log.Info("push-metrics-dump", "count", dump)
		}
		storePushQPS.Incr(int64(pack.Len))
	}
	dataLen, _ := strconv.ParseInt(r.Header.Get("Data-Length"), 10, 64)
	rw.WriteHeader(204)
	log.Debug("m-recv-ok",
		"len", dataLen,
		"bytes", len(pack.Data),
		"remote", r.RemoteAddr)
	metricsRecvQPS.Incr(dataLen)
}

func getPopSize(r *http.Request) int {
	size := 30
	if sizeStr := r.FormValue("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil {
			size = s
		}
	}
	return size
}

func metricsPop(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	metricsRopQPS.Incr(1)
	if mQueue == nil {
		httputil.Response404(rw, r)
		return
	}
	// pop value
	size := getPopSize(r)
	packets, err := mQueue.Pop(size)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		storePopFailDiff.Incr(1)
		return
	}
	if len(packets) == 0 {
		rw.WriteHeader(204)
		log.Debug("m-pop-0", "r", r.RemoteAddr)
		metricsRopZeroQPS.Incr(1)
		return
	}
	storePopAvg.Set(int64(len(packets)))

	// encode response
	rw.Header().Set("Data-Type", "pack")
	rw.Header().Set("Data-Length", strconv.Itoa(len(packets)))
	bytesLen, err := httputil.ResponseJSON(rw, packets, false, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		storePopFailDiff.Incr(1)
		return
	}
	log.Debug("m-pop-ok", "size", len(packets), "bytes", bytesLen, "r", r.RemoteAddr)

	// stats
	storeQueueLengthCount.Set(int64(mQueue.Len()))
	metricsPopDataQPS.Incr(int64(packets.DataLen()))
}

func metricsPopOld(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	metricsRopQPS.Incr(1)
	if mQueue == nil {
		httputil.Response404(rw, r)
		return
	}
	// pop value
	size := getPopSize(r)
	packets, err := mQueue.Pop(size)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		storePopFailDiff.Incr(1)
		return
	}
	if len(packets) == 0 {
		rw.WriteHeader(204)
		log.Debug("m-pop-0", "r", r.RemoteAddr)
		metricsRopZeroQPS.Incr(1)
		return
	}
	storePopAvg.Set(int64(len(packets)))

	// encode to old slices
	metrics, err := packets.ToMetricsList()
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		storePopFailDiff.Incr(1)
		return
	}
	rw.Header().Set("Data-Length", strconv.Itoa(len(metrics)))
	// old need slice slices
	bytesLen, err := httputil.ResponseJSON(rw, metrics, true, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		storePopFailDiff.Incr(1)
		return
	}
	log.Debug("m-pop-ok", "size", len(packets), "bytes", bytesLen, "r", r.RemoteAddr)

	// stats
	storeQueueLengthCount.Set(int64(mQueue.Len()))
	metricsPopDataQPS.Incr(int64(packets.DataLen()))
}

func openPing(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	metricsOpenReqQPS.Incr(1)
	users := getVerifyUsers(ps)
	userInfo := httptoken.GetUserVerifier(users["user"].(string))
	encoder := json.NewEncoder(rw)
	if err := encoder.Encode(userInfo); err != nil {
		httputil.ResponseErrorJSON(rw, r, 500, err)
		return
	}
	log.Debug("open-ping", "remote", httputil.RealIP(r), "tokens", users)
}

func openMetricRecv(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	metricsOpenReqQPS.Incr(1)
	users := getVerifyUsers(ps)
	pack, err := httputil.LoadPack(r, 1024*3)
	if err != nil {
		httputil.ResponseErrorJSON(rw, r, 500, err)
		log.Warn("open-m-recv-error", "remote", httputil.RealIP(r), "tokens", users, "error", err)
		return
	}
	if r.Form.Get("v2") == "" {
		pack.Data, err = convertOldMetric(pack.Data)
		if err != nil {
			httputil.ResponseErrorJSON(rw, r, 400, err)
			log.Warn("open-m-recv-error", "remote", httputil.RealIP(r), "tokens", users, "error", err)
			return
		}
	}
	if mQueue != nil {
		dump, ok := mQueue.Push(*pack)
		if !ok {
			httputil.ResponseFail(rw, r, ErrMetricsPushFail)
			log.Warn("open-m-recv-error", "err", err, "remote", httputil.RealIP(r))
			return
		}
		if dump > 0 {
			log.Info("open-push-dump", "count", dump)
		}
	}
	rw.WriteHeader(204)
	log.Debug("open-m-recv-ok",
		"bytes", len(pack.Data),
		"remote", httputil.RealIP(r),
		"store", r.Form.Get("store") != "",
		"v1", r.Form.Get("v2") == "",
		"user", users["user"])
	metricsOpenRecvQPS.Incr(int64(len(pack.Data) / 275)) // 275 is common size of metric in usage
}

func getVerifyUsers(ps httprouter.Params) map[string]interface{} {
	return map[string]interface{}{
		"user":  ps.ByName("user"),
		"token": ps.ByName("token"),
	}
}

func convertOldMetric(data []byte) ([]byte, error) {
	var oldMs []*models.MetricRaw
	if err := json.Unmarshal(data, &oldMs); err != nil {
		return nil, err
	}
	if len(oldMs) == 0 {
		return nil, errors.New("empty-metric-slice")
	}
	ms := make([]*models.Metric, 0, len(oldMs))
	for _, m := range oldMs {
		m2, err := m.ToNew()
		if err != nil {
			return nil, err
		}
		ms = append(ms, m2)
	}
	return json.Marshal(ms)
}

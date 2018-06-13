package transferhandler

import (
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/dataflow/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/julienschmidt/httprouter"
)

var (
	evtQueue *queues.Queue

	eventReqQPS  = expvar.NewQPS("http.event_req")
	eventRecvQPS = expvar.NewQPS("http.event_recv")
)

func init() {
	expvar.Register(eventReqQPS, eventRecvQPS)
}

func eventsRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	eventReqQPS.Incr(1)
	pack, err := httputil.LoadPack(r, 1024*3)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	if evtQueue != nil {
		dump, ok := evtQueue.Push(*pack)
		if !ok {
			httputil.ResponseFail(rw, r, ErrMetricsPushFail)
			log.Warn("e-recv-error", "err", err, "remote", r.RemoteAddr)
			return
		}
		if dump > 0 {
			log.Info("push-events-dump", "count", dump)
		}
	}
	rw.WriteHeader(204)
	dataLen, _ := strconv.ParseInt(r.Header.Get("Data-Length"), 10, 64)
	log.Debug("e-recv-ok",
		"len", dataLen,
		"gzip", r.Header.Get("Content-Length"),
		"remote", r.RemoteAddr)
	eventRecvQPS.Incr(dataLen)
}

package transferhandler

import (
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/julienschmidt/httprouter"
)

var (
	evtQueue *queues.Queue

	eventReqQPS     = expvar.NewQPS("http.event_req")
	eventRecvQPS    = expvar.NewQPS("http.event_recv")
	eventorPushQPS  = expvar.NewQPS("eventor.push")
	eventorDropDiff = expvar.NewDiff("eventor.drop")

	eventsFixLength int64 = 150
)

func init() {
	expvar.Register(eventReqQPS, eventRecvQPS,
		eventorPushQPS, eventorDropDiff)
}

func eventsRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	eventReqQPS.Incr(1)
	pack, err := httputil.LoadPack(r, 1024)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	if evtQueue != nil {
		dump, ok := evtQueue.Push(pack)
		if !ok {
			httputil.ResponseFail(rw, r, ErrMetricsPushFail)
			log.Warn("e-recv-error", "err", err, "remote", r.RemoteAddr)
			eventorDropDiff.Incr(int64(pack.Len))
			return
		}
		if dump > 0 {
			log.Info("push-events-dump", "count", dump)
		}
		eventorPushQPS.Incr(int64(pack.Len))
	}
	rw.WriteHeader(204)
	dataLen, _ := strconv.ParseInt(r.Header.Get("Data-Length"), 10, 64)
	if dataLen < 1 {
		dataLen = int64(len(pack.Data)) / eventsFixLength
		if dataLen < 1 {
			dataLen = 1
		}
	}
	log.Debug("e-recv-ok",
		"len", dataLen,
		"bytes", len(pack.Data),
		"remote", r.RemoteAddr)
	eventRecvQPS.Incr(dataLen)
}

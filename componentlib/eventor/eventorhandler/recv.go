package eventorhandler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/eventor/eventdata"
	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/julienschmidt/httprouter"
)

var (
	recvReqCount    = expvar.NewQPS("http.recv_req")
	recvEventsCount = expvar.NewQPS("http.recv_events")
)

func init() {
	expvar.Register(recvEventsCount, recvReqCount)
}

var (
	// ErrEventsZero means received zero events data
	ErrEventsZero = errors.New("events-0")
)

func eventsRecv(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	recvReqCount.Incr(1)
	var events [][]*models.Event
	if err := httputil.LoadJSON(r, &events); err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	if len(events) == 0 {
		httputil.ResponseFail(rw, r, ErrEventsZero)
		return
	}
	var count int
	for _, es := range events {
		count += len(es)
		go eventdata.Receive(es)
	}
	recvEventsCount.Incr(int64(count))
	// log.Debug("recv-ok", "count", count, "r", r.RemoteAddr)
	rw.WriteHeader(204)
}

func eventsRecvPacks(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	recvReqCount.Incr(1)
	dataLen, _ := strconv.Atoi(r.Header.Get("Data-Length"))
	packs, err := queues.PacketsFromReader(r.Body, dataLen)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	events, err := packs.ToEvents()
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	go eventdata.Receive(events)
	recvEventsCount.Incr(int64(dataLen))
	// log.Debug("recv-ok", "count", count, "r", r.RemoteAddr)
	rw.WriteHeader(204)
}

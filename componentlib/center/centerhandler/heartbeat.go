package centerhandler

import (
	"errors"
	"net/http"

	"github.com/baishancloud/mallard/componentlib/center/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/julienschmidt/httprouter"
)

var (
	reqHeartbeatCount      = expvar.NewDiff("http.req_heartbeat")
	reqHeartbeatHostsCount = expvar.NewAverage("http.req_heartbeat_hosts", 20)
)

func init() {
	expvar.Register(reqHeartbeatCount, reqHeartbeatHostsCount)
}

func heartbeatHandler(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqHeartbeatCount.Incr(1)
	heartbeats := make(map[string]models.EndpointHeartbeat)
	if err := httputil.LoadJSON(r, &heartbeats); err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	sqldata.UpdateHeartbeat(heartbeats)
	rw.WriteHeader(204)
	reqHeartbeatHostsCount.Set(int64(len(heartbeats)))
	log.Debug("req-heartbeat", "hosts", len(heartbeats), "r", r.RemoteAddr)
}

func hostServiceHandler(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqHeartbeatCount.Incr(1)
	svc := new(models.HostService)
	if err := httputil.LoadJSON(r, &svc); err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	if len(svc.Key()) < 2 {
		httputil.ResponseFail(rw, r, errors.New("wrong-hostservice"))
		return
	}
	sqldata.UpdateHostService(svc, r.RemoteAddr)
	rw.WriteHeader(204)
	log.Debug("req-hostservice", "key", svc.Key(), "r", r.RemoteAddr)
}

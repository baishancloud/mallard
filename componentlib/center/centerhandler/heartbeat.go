package centerhandler

import (
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

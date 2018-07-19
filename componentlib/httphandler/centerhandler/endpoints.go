package centerhandler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/compute/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/julienschmidt/httprouter"
)

var (
	reqEndpointCount = expvar.NewDiff("http.req_endpoint")
)

func init() {
	expvar.Register(reqEndpointCount)
}

func endpointsAllData(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	endpoints := sqldata.EndpointsAll()
	if endpoints == nil {
		httputil.Response404(rw, r)
		return
	}
	hash := r.FormValue("crc")
	if hash == fmt.Sprint(endpoints.CRC) {
		httputil.Response304(rw, r)
		return
	}
	rw.Header().Set("Content-Crc", fmt.Sprint(endpoints.CRC))
	isGzip := r.FormValue("gzip") != ""
	dataLen, err := httputil.ResponseJSON(rw, endpoints, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-endpoint-all", "r", r.RemoteAddr, "hash", endpoints.CRC, "bytes", dataLen, "is_gzip", isGzip)
}

func endpointsOneData(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	endpoint, hash := r.FormValue("endpoint"), r.FormValue("hash")
	if endpoint == "" {
		endpoint = r.FormValue("ep")
	}
	if endpoint == "" {
		httputil.ResponseFail(rw, r, errors.New("bad-endpoint"))
		return
	}
	epData := sqldata.EndpointOne(endpoint)
	if epData == nil {
		httputil.Response404(rw, r)
		return
	}
	if epData.Hash() == hash {
		httputil.Response304(rw, r)
		log.Debug("endpoint-one-304", "ep", endpoint, "r", r.RemoteAddr, "hash", hash)
		return
	}
	m := map[string]interface{}{
		"config": epData,
		"hash":   epData.Hash(),
	}
	rw.Header().Add("Data-Md5", epData.Hash())
	isGzip := (r.FormValue("gzip") == "1")
	dataLen, err := httputil.ResponseJSON(rw, m, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("endpoint-one", "ep", endpoint, "r", r.RemoteAddr, "data", dataLen, "rhash", hash, "hash", epData.Hash(), "is_gzip", isGzip)
}

type endpointSyncItem struct {
	Endpoint string `json:"endpoint"`
	Hash     string `json:"hash"`
}

func endpointSync(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	var items []endpointSyncItem
	if err := httputil.LoadJSON(r, &items); err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	if len(items) == 0 {
		httputil.ResponseFail(rw, r, errors.New("0-items"))
		return
	}
	endpoints := sqldata.EndpointsAll()
	if endpoints == nil {
		httputil.Response404(rw, r)
		return
	}

	var (
		changes []string
		deletes []string
	)
	for _, item := range items {
		epData := endpoints.Endpoint(item.Endpoint)
		if epData == nil {
			deletes = append(deletes, item.Endpoint)
			continue
		}
		if epData.Hash() != item.Hash {
			changes = append(changes, item.Endpoint)
		}
	}
	if len(changes) == 0 && len(deletes) == 0 {
		httputil.Response304(rw, r)
		return
	}

	m := map[string][]string{
		"updated": changes,
		"deleted": deletes,
	}
	httputil.ResponseJSON(rw, m, r.FormValue("gzip") == "1", false)
	log.Debug("endpoints-sync",
		"endpoints", len(items),
		"updates", len(m["updated"]),
		"deletes", len(m["deleted"]),
		"remote", r.RemoteAddr)
}

func endpointsLive(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	duration, _ := strconv.ParseInt(r.FormValue("duration"), 10, 64)
	timeRange, _ := strconv.ParseInt(r.FormValue("range"), 10, 64)
	if duration < 30 {
		duration = 30
	}
	if timeRange < 60 {
		timeRange = 60
	}
	endpoints, err := sqldata.NoLiveEndpoints(duration, timeRange)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	m := map[string]interface{}{"endpoints": endpoints}
	httputil.ResponseJSON(rw, m, false, false)
	log.Debug("endpoints-live", "r", r.RemoteAddr, "duration", duration, "timerange", timeRange, "endpoints", len(endpoints))
}

func endpointLiveAt(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	ep := r.FormValue("endpoint")
	if ep == "" {
		ep = r.FormValue("ep")
		if ep == "" {
			httputil.ResponseFail(rw, r, errors.New("bad-request"))
			return
		}
	}
	liveAt, err := sqldata.GetEndpointHeartbeatLive(ep)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	m := map[string]interface{}{
		"live_at":  liveAt,
		"endpoint": ep,
	}
	httputil.ResponseJSON(rw, m, false, false)
	log.Debug("heartbeat-live-at", "r", r.RemoteAddr, "endpoint", ep, "live_at", liveAt)
}

func endpointMaintain(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	data := sqldata.GetEndpointsMaintain()
	if data == nil {
		httputil.Response404(rw, r)
		return
	}
	httputil.ResponseJSON(rw, data, false, false)
	log.Debug("req-maintain-ok", "r", r.RemoteAddr, "maintains", len(data))
}

func endpointsInfos(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	infos := sqldata.HostInfosAll()
	if len(infos) == 0 {
		httputil.Response404(rw, r)
		return
	}
	isGzip := (r.FormValue("gzip") == "1")
	httputil.ResponseJSON(rw, infos, isGzip, false)
	log.Debug("req-host-infos-ok", "r", r.RemoteAddr, "is_gzip", isGzip)
}

func endpointsOneInfo(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqEndpointCount.Incr(1)
	ep := r.FormValue("endpoint")
	if ep == "" {
		ep = r.FormValue("ep")
		if ep == "" {
			httputil.ResponseFail(rw, r, errors.New("bad-request"))
			return
		}
	}
	infos := sqldata.HostInfosAll()
	if len(infos) == 0 {
		httputil.Response404(rw, r)
		return
	}
	isGzip := (r.FormValue("gzip") == "1")
	httputil.ResponseJSON(rw, infos[ep], isGzip, false)
	log.Debug("req-host-info-ok", "r", r.RemoteAddr, "ep", ep, "is_gzip", isGzip)
}

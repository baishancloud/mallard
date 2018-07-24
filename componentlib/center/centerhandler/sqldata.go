package centerhandler

import (
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/center/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/julienschmidt/httprouter"
)

var (
	reqStrategyCount    = expvar.NewDiff("http.req_startegy")
	reqTemplateCount    = expvar.NewDiff("http.req_template")
	reqGroupPluginCount = expvar.NewDiff("http.req_groupplugin")
)

func init() {
	expvar.Register(reqStrategyCount, reqTemplateCount, reqGroupPluginCount)
}

func strategyData(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqStrategyCount.Incr(1)
	dataHash := sqldata.DataHash()
	idStr := r.FormValue("id")
	isGzip := r.FormValue("gzip") != ""
	if idStr == "" {
		hash := r.FormValue("hash")
		if hash == dataHash {
			httputil.Response304(rw, r)
			return
		}
	}
	strategies := sqldata.StrategiesAll()
	if len(strategies) == 0 {
		httputil.Response404(rw, r)
		return
	}
	if idStr != "" {
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			httputil.Response404(rw, r)
			return
		}
		ss := strategies[id]
		if ss == nil {
			httputil.Response404(rw, r)
			return
		}
		dataLen, err := httputil.ResponseJSON(rw, ss, isGzip, false)
		if err != nil {
			httputil.ResponseFail(rw, r, err)
			return
		}
		log.Debug("req-strategies-one", "r", r.RemoteAddr, "id", idStr, "bytes", dataLen, "is_gzip", isGzip)
		return
	}
	rw.Header().Set("Content-Hash", dataHash)
	dataLen, err := httputil.ResponseJSON(rw, strategies, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-strategies-all", "r", r.RemoteAddr, "hash", dataHash, "bytes", dataLen, "is_gzip", isGzip)
}

func expressData(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqStrategyCount.Incr(1)
	dataHash := sqldata.DataHash()
	idStr := r.FormValue("id")
	isGzip := r.FormValue("gzip") != ""
	if idStr == "" {
		hash := r.FormValue("hash")
		if hash == dataHash {
			httputil.Response304(rw, r)
			return
		}
	}
	exps := sqldata.ExpressionsAll()
	if len(exps) == 0 {
		httputil.Response404(rw, r)
		return
	}
	if idStr != "" {
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			httputil.Response404(rw, r)
			return
		}
		ss := exps[id]
		if ss == nil {
			httputil.Response404(rw, r)
			return
		}
		dataLen, err := httputil.ResponseJSON(rw, ss, isGzip, false)
		if err != nil {
			httputil.ResponseFail(rw, r, err)
			return
		}
		log.Debug("req-expression-one", "r", r.RemoteAddr, "id", idStr, "bytes", dataLen, "is_gzip", isGzip)
		return
	}
	rw.Header().Set("Content-Hash", dataHash)
	dataLen, err := httputil.ResponseJSON(rw, exps, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-expression-all", "r", r.RemoteAddr, "hash", dataHash, "bytes", dataLen, "is_gzip", isGzip)
}

func templateData(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqTemplateCount.Incr(1)
	dataHash := sqldata.DataHash()
	isGzip := r.FormValue("gzip") != ""
	hash := r.FormValue("hash")
	idStr := r.FormValue("id")
	if idStr == "" {
		if hash == dataHash {
			httputil.Response304(rw, r)
			return
		}
	}
	tpls := sqldata.TemplatesAll()
	if len(tpls) == 0 {
		httputil.Response404(rw, r)
		return
	}
	if idStr != "" {
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			httputil.Response404(rw, r)
			return
		}
		tpl := tpls[id]
		if tpl == nil {
			httputil.Response404(rw, r)
			return
		}
		dataLen, err := httputil.ResponseJSON(rw, tpl, isGzip, false)
		if err != nil {
			httputil.ResponseFail(rw, r, err)
			return
		}
		log.Debug("req-template-one", "r", r.RemoteAddr, "id", idStr, "bytes", dataLen, "is_gzip", isGzip)
		return
	}
	rw.Header().Set("Content-Hash", dataHash)
	dataLen, err := httputil.ResponseJSON(rw, tpls, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-templates-all", "r", r.RemoteAddr, "hash", dataHash, "bytes", dataLen, "is_gzip", isGzip)
}

func groupPluginsData(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqGroupPluginCount.Incr(1)
	dataHash := sqldata.DataHash()
	hash := r.FormValue("hash")
	if hash == dataHash {
		httputil.Response304(rw, r)
		return
	}
	plugins := sqldata.GroupPluginsAll()
	if len(plugins) == 0 {
		httputil.Response404(rw, r)
		return
	}
	rw.Header().Set("Content-Hash", dataHash)
	isGzip := r.FormValue("gzip") != ""
	dataLen, err := httputil.ResponseJSON(rw, plugins, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-grouplugins-all", "r", r.RemoteAddr, "hash", dataHash, "bytes", dataLen, "is_gzip", isGzip)
}

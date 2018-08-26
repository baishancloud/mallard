package centerhandler

import (
	"net/http"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/pprofwrap"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/julienschmidt/httprouter"
)

var (
	log = zaplog.Zap("http")
)

// Handlers returns http handlers for center
func Handlers() http.Handler {
	r := httprouter.New()

	r.GET("/api/endpoints", endpointsAllData)
	r.GET("/api/endpoints/info", endpointsInfos)
	r.GET("/api/endpoint", endpointsOneData)
	r.POST("/api/endpoint/sync", endpointSync)
	r.GET("/api/endpoint/live", endpointsLive)
	r.GET("/api/endpoint/live_at", endpointLiveAt)
	r.GET("/api/endpoint/maintain", endpointMaintain)
	r.GET("/api/endpoint/info", endpointsOneInfo)

	r.GET("/api/host/services", hostServicesList)
	r.GET("/api/host/configs", hostConfigsList)

	r.GET("/api/alarm", alarmsAllData)
	r.GET("/api/alarm/all", alarmsAllData)
	r.GET("/api/alarm/request", alarmsOneRequest)
	r.GET("/api/alarm/for_strategy", alarmsOneForStrategy)
	r.GET("/api/alarm/requests", alarmsAllRequests)
	r.GET("/api/alarm/team", alarmsTeam)

	r.GET("/api/strategy", strategyData)
	r.GET("/api/expression", expressData)
	r.GET("/api/template", templateData)
	r.GET("/api/group_plugin", groupPluginsData)

	r.POST("/api/ping", heartbeatHandler)
	r.POST("/api/ping/hostservice", hostServiceHandler)

	r.HandlerFunc("GET", "/debug/vars", expvar.HTTPHandler)

	r.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		httputil.Response404(rw, r)
	})
	r = pprofwrap.Wrap(r)
	return r
}

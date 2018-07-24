package transferhandler

import (
	"net/http"

	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/pprofwrap"
	"github.com/julienschmidt/httprouter"
)

var (
	log = zaplog.Zap("http")
)

// Create creates transfer handler
func Create(isPublic bool) http.Handler {
	r := httprouter.New()
	r.POST("/api/metric", buildAuthorized(metricsRecv))
	r.GET("/api/config", (configGet))
	r.GET("/api/metric_pop", buildAuthorized(metricsPopOld))
	r.GET("/api/metric/pop", buildAuthorized(metricsPop))
	r.POST("/api/event", buildAuthorized(eventsRecv))
	/*r.POST("/api/agentself", s.buildAuthorized(s.selfInfo))*/

	if isPublic {
		r.GET("/open/ping", buildVerifier(openPing))
		r.POST("/open/metric", buildVerifier(openMetricRecv))
	}

	r.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		httputil.Response404(rw, r)
	})
	r.HandlerFunc("GET", "/debug/vars", expvar.HTTPHandler)
	r = pprofwrap.Wrap(r)
	return r
}

func buildAuthorized(handler httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !httptoken.CheckHeaderResponse(rw, r) {
			httputil.Response401(rw, r)
			return
		}
		handler(rw, r, ps)
	}
}

// SetQueues sets queues for server
func SetQueues(mQ, eQ *queues.Queue) {
	mQueue = mQ
	evtQueue = eQ
}

func buildVerifier(handler httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user, token, err := httptoken.VerifyAndAllow(r)
		if err == nil {
			ps = append(ps, httprouter.Param{
				Key:   "user",
				Value: user,
			}, httprouter.Param{
				Key:   "token",
				Value: token,
			})
			handler(rw, r, ps)
			return
		}
		addon := map[string]interface{}{
			"user":  user,
			"token": token,
		}
		if err == httptoken.ErrorTokenInvalid {
			httputil.Response401(rw, r)
			log.Warn("verify-invalid", "tokens", addon)
			return
		}
		if err == httptoken.ErrorLimitExceeded {
			rw.WriteHeader(http.StatusTooManyRequests)
			log.Warn("verify-limited", "tokens", addon)
			return
		}
		httputil.ResponseFail(rw, r, nil)
		log.Warn("verify-fail", "tokens", addon)
	}
}

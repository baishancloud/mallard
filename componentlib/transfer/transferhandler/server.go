package transferhandler

import (
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/pprofwrap"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/configapi"
	"github.com/julienschmidt/httprouter"
)

var (
	log = zaplog.Zap("http")
)

// Create creates transfer handler
func Create(isPublic bool, isAuthorized bool) http.Handler {
	r := httprouter.New()
	r.POST("/api/metric", buildAuthorized(metricsRecv, isAuthorized))
	r.POST("/api/event", buildAuthorized(eventsRecv, isAuthorized))
	r.GET("/api/config", buildAuthorized(configGet, isAuthorized))

	r.GET("/api/metric_pop", buildAuthorized(metricsPopOld, isAuthorized))
	r.GET("/api/metric_pop2", buildAuthorized(metricsPop, isAuthorized))
	r.POST("/api/selfinfo", buildAuthorized(nil, isAuthorized))

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

var (
	recvIgnoreQPS = expvar.NewQPS("http.recv_ignore")
)

func init() {
	expvar.Register(recvIgnoreQPS)
}

func buildAuthorized(handler httprouter.Handle, isAuthorized bool) httprouter.Handle {
	if !isAuthorized {
		return handler
	}
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !httptoken.CheckHeaderResponse(rw, r) {
			httputil.Response401(rw, r)
			return
		}
		// ignore data request for ignored agent, except /api/config
		if ep := r.Header.Get("Agent-Endpoint"); ep != "" && r.Header.Get("Data-Length") != "" {
			if configapi.CheckAgentStatus(ep, models.AgentStatusIgnore) {
				dataLen, _ := strconv.ParseInt(r.Header.Get("Data-Length"), 10, 64)
				rw.WriteHeader(204)
				log.Debug("recv-ignore",
					"endpoint", ep,
					"len", dataLen,
					"remote", r.RemoteAddr)
				recvIgnoreQPS.Incr(dataLen)
				return
			}
		}
		if handler != nil {
			handler(rw, r, ps)
		}
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

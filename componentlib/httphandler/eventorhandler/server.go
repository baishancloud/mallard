package eventorhandler

import (
	"net/http"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/pprofwrap"
	"github.com/julienschmidt/httprouter"
)

var (
	log = zaplog.Zap("http")
)

// Create creates transfer handler
func Create() http.Handler {
	r := httprouter.New()
	r.POST("/push/event", eventsRecv)
	r.POST("/api/event", eventsRecvPacks)
	r.HandlerFunc("GET", "/debug/vars", expvar.HTTPHandler)

	r.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		httputil.Response404(rw, r)
	})
	r = pprofwrap.Wrap(r)
	return r
}

package judgehandler

import (
	"net/http"

	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/pprofwrap"
	"github.com/julienschmidt/httprouter"
)

var (
	log = zaplog.Zap("http")
)

var (
	queue chan<- []*models.Metric
)

// SetQueue sets metric channel
func SetQueue(mQueue chan<- []*models.Metric) {
	queue = mQueue
}

// Create creates http handlers
func Create() http.Handler {
	r := httprouter.New()
	r.POST("/api/metric", metricsRecv)
	r.POST("/write", influxRecv)
	//r.GET("/api/query", s.metricQuery)

	r.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		httputil.Response404(rw, r)
	})
	r = pprofwrap.Wrap(r)
	return r
}

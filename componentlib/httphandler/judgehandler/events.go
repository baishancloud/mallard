package judgehandler

import (
	"net/http"

	"github.com/baishancloud/mallard/componentlib/compute/multijudge"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/julienschmidt/httprouter"
)

func judgeEvents(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	events := multijudge.AllCachedEvents()
	httputil.ResponseJSON(rw, events, false, false)
}

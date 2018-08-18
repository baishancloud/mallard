package transferhandler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/extralib/configapi"
	"github.com/julienschmidt/httprouter"
)

var (
	configReqQPS = expvar.NewQPS("http.config_req")
)

func init() {
	expvar.Register(configReqQPS)
}

func configGet(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	configReqQPS.Incr(1)

	rw.Header().Set("Transfer-Time", strconv.FormatInt(time.Now().Unix(), 10))
	endpoint, hash := r.FormValue("endpoint"), r.FormValue("hash")
	if endpoint == "" {
		endpoint = r.FormValue("ep") // try ep param
	}
	if endpoint == "" {
		httputil.ResponseFail(rw, r, errors.New("bad-params"))
		return
	}

	// set heart beat
	configapi.SetHeartbeat(endpoint,
		r.Header.Get("Agent-Version"),
		r.Header.Get("Agent-Plugin"),
		r.Header.Get("Agent-IP"),
		strings.Split(r.RemoteAddr, ":")[0],
	)

	epData := configapi.EndpointConfig(endpoint)
	if epData == nil {
		httputil.Response404(rw, r)
		return
	}
	if hash != "" && hash == epData.Hash() {
		rw.WriteHeader(304)
		log.Debug("config-get-304", "ep", endpoint, "hash", hash)
		return
	}
	hash = epData.Hash()
	mData := map[string]interface{}{
		"config": epData,
		"hash":   hash,
	}
	isGzip := (r.FormValue("gzip") != "")
	rw.Header().Set("Content-Hash", hash)
	rw.Header().Set("Transfer-Sertypes", configapi.GetEndpointSertypes(endpoint))
	httputil.ResponseJSON(rw, mData, isGzip, false)
	log.Debug("config-get-ok", "ep", endpoint, "hash", hash, "gzip", isGzip)
}

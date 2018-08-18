package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log = zaplog.Zap("http")

	reqOKQPS   = expvar.NewQPS("http.req_ok")
	reqFailQPS = expvar.NewQPS("http.req_fail")
	req404QPS  = expvar.NewQPS("http.req_404")
	req304QPS  = expvar.NewQPS("http.req_304")
	req401QPS  = expvar.NewQPS("http.req_401")

	svr *http.Server
)

func init() {
	expvar.Register(reqFailQPS, req401QPS, req404QPS, req304QPS, reqOKQPS)
}

// Listen listens http server
func Listen(addr string, handler http.Handler, certs ...string) {
	svr = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Second * 30,
		WriteTimeout:      time.Minute * 3,
	}
	if len(certs) == 2 && certs[0] != "" && certs[1] != "" {
		log.Info("init-https", "addr", addr)
		if err := svr.ListenAndServeTLS(certs[0], certs[1]); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			log.Fatal("listen-error", "error", err)
		}
		return
	}
	log.Info("init", "addr", addr)
	if err := svr.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			return
		}
		log.Fatal("listen-error", "error", err)
	}
}

// Close closes http server
func Close() {
	log.Info("close")
	if svr != nil {
		svr.Close()
	}
}

var (
	// ErrBodyEmpty means body is empty after reading
	ErrBodyEmpty = errors.New("empty-body")
	// ErrBodyWrongContentLength means body length is not matched with header Content-Length
	ErrBodyWrongContentLength = errors.New("content-length-wrong")
)

const (
	// ContentTypePack is raw pack bytes
	ContentTypePack = "application/m-pack"
	// ContentTypeGzipJSON is compressed json bytes
	ContentTypeGzipJSON = "application/gzip+json"
	// ContentTypeJSON is json bytes
	ContentTypeJSON = "application/json"
)

// LoadPack loads body to transfer packet
func LoadPack(r *http.Request, capacity int64) (*queues.Packet, error) {
	data, err := osutil.ReadAll(r.Body, capacity)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrBodyEmpty
	}
	contentLength := r.Header.Get("Content-Length")
	if contentLength != "" && contentLength != strconv.Itoa(len(data)) {
		return nil, ErrBodyWrongContentLength
	}
	pack := &queues.Packet{
		Data: data,
	}
	dataLen, _ := strconv.Atoi(r.Header.Get("Data-Length"))
	pack.Len = dataLen
	if r.Header.Get("Content-Type") == ContentTypeGzipJSON {
		pack.Type = queues.PacketTypeGzip
	}
	return pack, nil
}

// LoadJSON loads body to json object
func LoadJSON(r *http.Request, v interface{}) error {
	if r.Header.Get("Content-Type") == ContentTypeGzipJSON {
		return utils.UngzipJSON(r.Body, v)
	}
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// ResponseFail responses fial
func ResponseFail(rw http.ResponseWriter, r *http.Request, err error) {
	rw.WriteHeader(500)
	rw.Write([]byte(err.Error()))
	log.Warn("500", "u", r.RequestURI, "r", r.RemoteAddr, "error", err)
	reqFailQPS.Incr(1)
}

// Response404 responses 404
func Response404(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(404)
	rw.Write([]byte(http.StatusText(http.StatusNotFound)))
	log.Warn("404", "u", r.RequestURI, "r", r.RemoteAddr)
	req404QPS.Incr(1)
}

// Response401 responses 401
func Response401(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(401)
	rw.Write([]byte(http.StatusText(http.StatusUnauthorized)))
	log.Warn("401", "u", r.RequestURI, "r", r.RemoteAddr)
	req401QPS.Incr(1)
}

// Response304 responses 304
func Response304(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(304)
	log.Warn("304", "u", r.RequestURI, "r", r.RemoteAddr)
	req304QPS.Incr(1)
}

// ResponseJSON writes json data to response
func ResponseJSON(rw http.ResponseWriter, value interface{}, isGzip bool, withMD5 bool) (int, error) {
	n, err := responseJSON(rw, 200, value, isGzip, withMD5)
	if err != nil {
		return 0, err
	}
	reqOKQPS.Incr(1)
	return n, nil
}

// ResponseOkJSON responses json data
func ResponseOkJSON(rw http.ResponseWriter, value interface{}) error {
	_, err := responseJSON(rw, 200, map[string]interface{}{
		"status": "ok",
		"data":   value,
	}, false, false)
	if err != nil {
		return err
	}
	reqOKQPS.Incr(1)
	return nil
}

// ResponseErrorJSON response error json
func ResponseErrorJSON(rw http.ResponseWriter, r *http.Request, status int, err error) {
	responseJSON(rw, status, map[string]interface{}{
		"status": "error",
		"error":  err.Error(),
	}, false, false)
	log.Warn("fail", "u", r.RequestURI, "r", r.RemoteAddr, "error", err)
	reqFailQPS.Incr(1)
}

func responseJSON(rw http.ResponseWriter, status int, value interface{}, isGzip bool, withMD5 bool) (int, error) {
	var (
		data []byte
		err  error
	)
	if isGzip {
		data, err = utils.GzipJSONBytes(value, 1024*10)
		rw.Header().Set("Content-Type", ContentTypeGzipJSON)
	} else {
		data, err = json.Marshal(value)
		rw.Header().Set("Content-Type", ContentTypeJSON)
	}
	if err != nil {
		return 0, err
	}
	if withMD5 {
		rw.Header().Set("Data-Md5", utils.MD5HashBytes(data))
	}
	rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
	rw.WriteHeader(status)
	rw.Write(data)
	return len(data), nil
}

// RealIP gets real ip from request
func RealIP(r *http.Request) string {
	if r.Header.Get("X-Real-IP") != "" {
		return r.Header.Get("X-Real-IP")
	}
	return r.RemoteAddr
}

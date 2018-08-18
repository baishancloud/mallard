package transfer

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	log.SetDebug(true)
}

func setupTransferServer(c C) func() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", func(rw http.ResponseWriter, r *http.Request) {
		c.So(r.Header.Get("Agent-Version"), ShouldEqual, "version")
		c.So(r.Header.Get("Agent-Buildtime"), ShouldEqual, "build-time")
		rw.Header().Set("Transfer-Time", strconv.FormatInt(time.Now().Unix(), 10))
		hash := r.FormValue("hash")
		if hash != "" && hash == "c75da60fc03cf89a4d3666420602a94d" {
			rw.WriteHeader(304)
			return
		}
		epData := &models.EndpointConfig{
			Builtin: &models.EndpointBuiltin{},
		}
		hash = epData.Hash()
		mData := map[string]interface{}{
			"config": epData,
			"hash":   hash,
		}
		isGzip := (r.FormValue("gzip") != "")
		httputil.ResponseJSON(rw, mData, isGzip, false)
	})
	mux.HandleFunc("/api/event", func(rw http.ResponseWriter, r *http.Request) {
		c.So(r.Header.Get("Data-Length"), ShouldEqual, "1500")
		rw.WriteHeader(204)
	})
	mux.HandleFunc("/api/metric", func(rw http.ResponseWriter, r *http.Request) {
		c.So(r.Header.Get("Hash-Key"), ShouldNotBeEmpty)
		c.So(r.Header.Get("Hash-Code"), ShouldNotBeEmpty)
		c.So(r.Header.Get("Data-Length"), ShouldEqual, "1500")
		rw.WriteHeader(204)
	})
	mux.HandleFunc("/api/selfinfo", func(rw http.ResponseWriter, r *http.Request) {
		c.So(r.Header.Get("Hash-Key"), ShouldNotBeEmpty)
		c.So(r.Header.Get("Hash-Code"), ShouldNotBeEmpty)
		postData := make(map[string]interface{})
		err := utils.UngzipJSON(r.Body, &postData)
		c.So(err, ShouldBeNil)
		c.So(postData, ShouldContainKey, "serverinfo")
		c.So(postData, ShouldContainKey, "endpoint")
		c.So(postData, ShouldContainKey, "config")
		c.So(postData, ShouldContainKey, "plugins")
		rw.WriteHeader(204)
	})
	server = httptest.NewServer(mux)

	return func() {
		server.Close()
	}
}

func setTestURLs() {
	SetURLs([]string{server.URL}, map[string]string{
		"metric": "/api/metric",
		"event":  "/api/event",
		"config": "/api/config",
		"self":   "/api/selfinfo",
	})
	serverinfo.Read("default-endpoint", true)
}

func TestTransfer(t *testing.T) {
	Convey("transfer", t, func(c C) {
		closeFn := setupTransferServer(c)
		setTestURLs()
		// test config
		requestConfig(SyncOption{
			Interval:  time.Minute,
			Version:   "version",
			BuildTime: "build-time",
			Func: func(data *models.EndpointData, isUpdate bool) {
				c.So(data.Time > 0, ShouldBeTrue)
				c.So(isUpdate, ShouldBeTrue)
				c.So(data.Hash, ShouldEqual, "c75da60fc03cf89a4d3666420602a94d")
			},
		})
		requestConfig(SyncOption{
			Interval:  time.Minute,
			Version:   "version",
			BuildTime: "build-time",
			Func: func(data *models.EndpointData, isUpdate bool) {
				c.So(data.Time > 0, ShouldBeTrue)
				c.So(isUpdate, ShouldBeFalse)
				c.So(data.Hash, ShouldEqual, "c75da60fc03cf89a4d3666420602a94d")
			},
		})

		Events(make([]*models.Event, 3000))
		Metrics(make([]*models.Metric, 3000))
		SendSelfInfo("self-config")
		closeFn()
	})
}

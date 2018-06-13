package transfer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/utils"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func setupServer() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	return func() {
		server.Close()
	}
}

func TestClient(t *testing.T) {
	Convey("client", t, func() {
		Convey("get", func(c C) {
			defer setupServer()()
			client := NewClient(time.Second, 10, "token")

			mux.HandleFunc("/test", func(rw http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, "GET")
				c.So(r.Header.Get("X-Header"), ShouldEqual, "abc")
				c.So(httptoken.CheckHeader(r.Header), ShouldBeTrue)
			})

			resp, du, err := client.GET(server.URL+"/test", map[string]string{
				"X-Header": "abc",
			})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(du.Nanoseconds() > 0, ShouldBeTrue)
			resp.Body.Close()
		})
		Convey("post", func(c C) {
			defer setupServer()()
			client := NewClient(time.Second, 10, "token")

			testData := strings.Repeat("a", 1e3)

			mux.HandleFunc("/test", func(rw http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, "POST")
				header := r.Header
				c.So(header.Get("Content-Type"), ShouldEqual, "application/gzip+json")
				c.So(header.Get("Data-Length"), ShouldEqual, "10")
				var str string
				err := utils.UngzipJSON(r.Body, &str)
				c.So(err, ShouldBeNil)
				c.So(str, ShouldEqual, testData)
				rw.WriteHeader(204)
			})

			resp, du, err := client.POST(server.URL+"/test", testData, 10)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.StatusCode, ShouldEqual, 204)
			So(du.Nanoseconds() > 0, ShouldBeTrue)
			resp.Body.Close()
		})
		Convey("post.400+", func(c C) {
			defer setupServer()()
			client := NewClient(time.Second, 10, "token")

			testData := strings.Repeat("a", 1e3)

			mux.HandleFunc("/test", func(rw http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, "POST")
				rw.WriteHeader(404)
			})

			resp, du, err := client.POST(server.URL+"/test", testData, 10)
			So(err, ShouldHaveSameTypeAs, ClientError{})
			clientErr := err.(ClientError)
			So(clientErr.Status, ShouldEqual, 404)
			So(resp, ShouldBeNil)
			So(du, ShouldEqual, 0)
		})
	})
}

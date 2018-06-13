package httptoken

import (
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthHeader(t *testing.T) {
	Convey("auth-header", t, func() {
		headers := BuildHeader("abc")
		So(headers, ShouldContainKey, KeyHeader)
		So(headers, ShouldContainKey, CodeHeader)

		httpHeader := make(map[string][]string)
		for k, v := range headers {
			httpHeader[k] = []string{v}
		}
		check := CheckHeader(httpHeader)
		So(check, ShouldBeTrue)

		Convey("time-out", func() {
			headers := BuildHeader("abc", time.Now().Unix()-HashDuration*2)
			httpHeader := make(map[string][]string)
			for k, v := range headers {
				httpHeader[k] = []string{v}
			}
			check := CheckHeader(httpHeader)
			So(check, ShouldBeFalse)

			headers = BuildHeader("abc", time.Now().Unix()+HashDuration*2)
			httpHeader = make(map[string][]string)
			for k, v := range headers {
				httpHeader[k] = []string{v}
			}
			check = CheckHeader(httpHeader)
			So(check, ShouldBeFalse)
		})
	})

	Convey("auth-http", t, func() {
		req := httptest.NewRequest("GET", "/", nil)
		headers := BuildHeader("abc")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rw := httptest.NewRecorder()
		check := CheckHeaderResponse(rw, req)
		So(check, ShouldBeTrue)

		req = httptest.NewRequest("GET", "/", nil)
		headers = BuildHeader("abc", time.Now().Unix()-HashDuration*2)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rw = httptest.NewRecorder()
		check = CheckHeaderResponse(rw, req)
		So(rw.Code, ShouldEqual, 401)
		So(rw.Header().Get("Connection"), ShouldEqual, "close")
	})
}

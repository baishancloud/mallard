package httptoken

import (
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestVerifier(t *testing.T) {
	Convey("verfier", t, func() {
		err := refreshVerifyFile("notfound.json")
		So(err, ShouldNotBeNil)
		So(tokenMap, ShouldHaveLength, 0)

		err = refreshVerifyFile("test_token.json")
		So(err, ShouldBeNil)

		So(tokenMap, ShouldContainKey, "abc")
		user := GetUserVerifier("abc")
		So(user.RateLimit, ShouldEqual, DefaultRateLimit)

		req := httptest.NewRequest("GET", "/?user="+user.User+"&token="+user.Token, nil)
		_, _, check := VerifyRequest(req)
		So(check, ShouldBeTrue)

		_, _, err = VerifyAndAllow(req)
		So(err, ShouldBeNil)

		check = VerifyAllowLimit("abc")
		So(check, ShouldBeTrue)
		for i := 0; i < 10; i++ {
			VerifyAllowLimit("abc")
		}
		_, _, err = VerifyAndAllow(req)
		So(err, ShouldEqual, ErrorLimitExceeded)

		req2 := httptest.NewRequest("GET", "/?user="+user.User+"&token="+user.Token+"xyz", nil)
		_, _, err = VerifyAndAllow(req2)
		So(err, ShouldEqual, ErrorTokenInvalid)

		err = refreshVerifyFile("test_token.json")
		So(err, ShouldBeNil)
	})
}

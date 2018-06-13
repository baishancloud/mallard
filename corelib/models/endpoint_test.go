package models

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEndpoint(t *testing.T) {
	Convey("endpoint", t, func() {
		ep := new(EndpointConfig)
		So(ep.Len(), ShouldEqual, 0)
		So(ep.Hash(), ShouldEqual, "99914b932bd37a50b983c5e7c90ae93b")
	})
}

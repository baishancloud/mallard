package msggcall

import (
	"strings"
	"testing"
	"time"

	"github.com/baishancloud/mallard/componentlib/eventor/redisdata"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	log = zaplog.Null()
}

var nowUnix = time.Now().Unix()
var testEventFull = &models.EventFull{
	ID: "s_123456",
	Strategy: &models.Strategy{
		Metric:   "cpu",
		Note:     "cpu-error",
		ID:       999,
		Priority: 4,
	},
	Status:    models.EventProblem.String(),
	EventTime: nowUnix,
	Endpoint:  "localhost",
	PushedTags: map[string]string{
		"sertypes": "node-server",
	},
	Fields: map[string]interface{}{
		"value2": 2.2,
	},
}

func TestSetFiles(t *testing.T) {
	Convey("SetFiles", t, func() {
		SetFiles("a.sh", "b.sh", "test.sh", "d.sh")
		So(commandFile, ShouldBeEmpty)
		So(actionFile, ShouldBeEmpty)
		So(msggFile, ShouldEqual, "test.sh")
		So(msggFileWay, ShouldBeEmpty)
	})
}

func TestCallEvent(t *testing.T) {
	Convey("CallEvent", t, func() {
		Convey("shouldCall", func() {
			st := getCallStrategy(redisdata.EventRecord{
				IsHigh: true,
				Event:  new(models.EventFull),
			})
			So(st, ShouldBeNil)

			st = getCallStrategy(redisdata.EventRecord{
				IsHigh: true,
				Event:  testEventFull,
			})
			So(st, ShouldNotBeNil)
			So(st.ID, ShouldEqual, 999)

			st = getCallStrategy(redisdata.EventRecord{
				IsHigh: false,
				Event:  testEventFull,
			})
			So(st, ShouldBeNil)
		})

		Convey("callCommand", func() {
			commandFile = "./test.sh"
			st := getCallStrategy(redisdata.EventRecord{
				IsHigh: true,
				Event:  testEventFull,
			})
			output, err := CallCommand(testEventFull, st.Note, "uic")
			So(err, ShouldBeNil)
			So(strings.Trim(output, "\n"), ShouldEqual, "localhost cpu-error 4 node-server uic 0.000")
		})

		Convey("callAction", func() {
			actionFile = "./test.sh"
			st := getCallStrategy(redisdata.EventRecord{
				IsHigh: true,
				Event:  testEventFull,
			})
			output, err := CallAction(testEventFull, st.Note, "uic")
			So(err, ShouldBeNil)
			So(strings.Trim(output, "\n"), ShouldEqual, `localhost cpu-error 4 {"sertypes":"node-server"} uic 0.000 {"value2":2.2}`)
		})

		Convey("callMsgg", func() {
			msggFile = "./test.sh"
			st := getCallStrategy(redisdata.EventRecord{
				IsHigh: true,
				Event:  testEventFull,
			})
			Convey("addRequest", func() {
				AddRequests(testEventFull, st, true)
				So(requests[testEventFull.ID], ShouldHaveLength, 1)
				req := requests[testEventFull.ID]
				So(req[nowUnix], ShouldNotBeNil)
				r := req[nowUnix]
				So(r.Note, ShouldEqual, st.Note)
				So(r.Endpoint, ShouldEqual, testEventFull.Endpoint)
				So(r.Level, ShouldEqual, 5-st.Priority)
				So(r.SendRequest.Uic, ShouldEqual, "000")
				So(r.SendRequest.Template, ShouldEqual, "000")
				So(r.SendRequest.Emails, ShouldHaveLength, 0)
				So(r.Recover, ShouldBeTrue)
			})
			Convey("cleanRequest", func() {
				CleanRequests(testEventFull.ID)
				So(requests[testEventFull.ID], ShouldHaveLength, 0)
			})
		})
	})
}

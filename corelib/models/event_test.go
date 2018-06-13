package models

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEvent(t *testing.T) {
	Convey("event", t, func() {
		event := &Event{
			ID:        "123",
			Status:    EventOk,
			Time:      time.Now().Unix(),
			Strategy:  123,
			Endpoint:  "localhost",
			LeftValue: 123.456,
			Step:      3,
			Fields: map[string]interface{}{
				"abc": 1234,
			},
			Tags: map[string]string{
				"cachegroup": "dx-fujian-xiamen-1-cache-1",
				"sertypes":   "cache|cache_edge",
			},
		}
		se := event.Simple()
		So(se.Fields, ShouldHaveLength, 0)
		So(se.LeftValue, ShouldEqual, 0)

		fullTags := event.FullTags()
		So(fullTags["cachegroup_province"], ShouldEqual, "fujian")
		So(fullTags["cachegroup_isp"], ShouldEqual, "dx")
		So(fullTags["cachegroup_city"], ShouldEqual, "xiamen")
		So(fullTags["sertypes_cache"], ShouldEqual, "1")
		So(fullTags["sertypes_cache_edge"], ShouldEqual, "1")

		Convey("event.full", func() {
			fullEvent := &EventFull{
				ID:        "123",
				Status:    EventOk.String(),
				EventTime: time.Now().Unix(),
				Strategy:  123,
				Endpoint:  "localhost",
				LeftValue: 123.456,
				Fields: map[string]interface{}{
					"abc": 1234,
				},
			}
			So(fullEvent.Priority(), ShouldEqual, 0)
			st := &Strategy{
				ID:       123,
				Priority: 6,
			}
			fullEvent.Strategy = st
			So(fullEvent.Priority(), ShouldEqual, 6)

			dto := NewEventDto(fullEvent, st)
			So(dto.Priority, ShouldEqual, 6)
		})
	})
}

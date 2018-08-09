package models

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAlarm(t *testing.T) {
	Convey("alarm", t, func() {
		_, offset := time.Now().Zone()
		Convey("user.status", func() {
			us := &UserStatus{
				EmailStatus: 1,
			}
			So(us.IsEnable(), ShouldBeTrue)
			So(us.Status(), ShouldResemble, []int{1, 0, 0, 0})

			us2 := &UserStatus{}
			So(us2.IsEnable(), ShouldBeFalse)
		})
		Convey("team.status", func() {
			tu := &TeamUserStatus{
				StartTime: 3600,
				EndTime:   7200,
			}
			So(tu.IsInTime(3890-int64(offset)), ShouldBeTrue)
			So(tu.IsInTime(7899-int64(offset)), ShouldBeFalse)

			tu2 := &TeamUserStatus{
				EndTime:   3600,
				StartTime: 7200,
			} // means 0-7200 ~ 86400+3600 ~ 86400*2-0
			So(tu2.IsInTime(3890-int64(offset)), ShouldBeFalse)
			So(tu2.IsInTime(7899-int64(offset)), ShouldBeTrue)
		})
		Convey("du.info", func() {
			tu := &DutyInfo{
				BeginTime: 3600,
				EndTime:   7200,
			}
			So(tu.IsInTime(3890-int64(offset)), ShouldBeTrue)
			So(tu.IsInTime(7899-int64(offset)), ShouldBeFalse)

			tu2 := &DutyInfo{
				EndTime:   3600,
				BeginTime: 7200,
			} // means 0-7200 ~ 86400+3600 ~ 86400*2-0
			So(tu2.IsInTime(3890-int64(offset)), ShouldBeFalse)
			So(tu2.IsInTime(7899-int64(offset)), ShouldBeTrue)

			tu3 := &DutyInfo{
				EndTime:   3600,
				BeginTime: 3600,
			} // means no time in range
			So(tu3.IsInTime(123), ShouldBeFalse)
		})
	})
}

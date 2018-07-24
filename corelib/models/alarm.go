package models

import (
	"time"
)

type (
	// UserInfo is user info
	UserInfo struct {
		ID     int    `db:"id" json:"id"`
		Name   string `db:"name" json:"name,omitempty"`
		Cnname string `db:"cnname" json:"cnname,omitempty"`
		Email  string `db:"email" json:"email,omitempty"`
		Phone  string `db:"phone" json:"phone,omitempty"`
		IM     string `db:"im" json:"im,omitempty"`
		QQ     string `db:"qq" json:"qq,omitempty"`
		Role   int    `db:"role" json:"role,omitempty"`
	}
	// UserStatus is alarm status for user and one strategy
	UserStatus struct {
		UserID       int    `db:"user_id" json:"user_id,omitempty"`
		StrategyID   int    `db:"strategy_id" json:"strategy_id,omitempty"`
		EmailStatus  int    `db:"email_status" json:"email,omitempty"`
		PhoneStatus  int    `db:"phone_status" json:"phone,omitempty"`
		WechatStatus int    `db:"wechat_status" json:"wechat,omitempty"`
		SmsStatus    int    `db:"sms_status" json:"sms,omitempty"`
		Step         int    `db:"step" json:"step,omitempty"`
		UserName     string `db:"-" json:"user_name,omitempty"`
		UserNameCN   string `db:"-" json:"user_name_cn,omitempty"`
	}
	// OuterUserInfo is info of outer users, not registered or employee
	OuterUserInfo struct {
		ID          int    `db:"id" json:"id"`
		Name        string `db:"name" json:"name,omitempty"`
		Email       string `db:"email" json:"email,omitempty"`
		Phone       string `db:"phone" json:"phone,omitempty"`
		CompanyID   int    `db:"company_id" json:"company_id,omitempty"`
		CompanyName string `db:"company_name" json:"company,omitempty"`
	}
)

// IsEnable checks the alarm status is enable,
// if no way is set, return false
func (a *UserStatus) IsEnable() bool {
	return a.EmailStatus+a.PhoneStatus+a.SmsStatus+a.WechatStatus > 0
}

// Status returns alarm status list from each way
func (a *UserStatus) Status() []int {
	return []int{a.EmailStatus, a.WechatStatus, a.SmsStatus, a.PhoneStatus}
}

type (
	// TeamStrategy is relation of team and one strategy
	TeamStrategy struct {
		TeamID     int `db:"team_id" json:"team_id"`
		StrategyID int `db:"strategy_id" json:"strategy_id"`
		Step       int `db:"step" json:"step,omitempty"`
	}
	// TeamInfo is info of one team
	TeamInfo struct {
		ID     int    `db:"id" json:"id"`
		Name   string `db:"name" json:"name,omitempty"`
		Title  string `db:"title" json:"title,omitempty"`
		Remark string `db:"remark" json:"remark,omitempty"`
	}
	// TeamUserStatus is alarm status for the all users in the team
	TeamUserStatus struct {
		UserID       int    `db:"user_id" json:"user_id"`
		AlarmTeamID  int    `db:"alarm_team_id" json:"alarm_team_id,omitempty"`
		StartTime    int    `db:"start_time" json:"start_time,omitempty"`
		EndTime      int    `db:"end_time" json:"end_time,omitempty"`
		EmailStatus  int    `db:"email_status" json:"email,omitempty"`
		PhoneStatus  int    `db:"phone_status" json:"phone,omitempty"`
		WechatStatus int    `db:"wechat_status" json:"wechat,omitempty"`
		SmsStatus    int    `db:"sms_status" json:"sms,omitempty"`
		Step         int    `db:"step" json:"step,omitempty"`
		UserName     string `db:"-" json:"user_name,omitempty"`
		UserNameCN   string `db:"-" json:"user_name_cn,omitempty"`
	}
)

// IsInTime checks whether the time is in range for the team to alarm
func (a *TeamUserStatus) IsInTime(t int64) bool {
	if a.StartTime == 0 || a.EndTime == 0 {
		return true
	}
	now := time.Unix(t, 0).Format("15:04")
	t1 := time.Unix(int64(a.StartTime), 0).In(time.UTC).Format("15:04")
	t2 := time.Unix(int64(a.EndTime), 0).In(time.UTC).Format("15:04")
	if a.EndTime > a.StartTime {
		return now >= t1 && now <= t2
	}
	return !(now >= t2 && now <= t1)
}

type (
	// DutyInfo is info for once duty
	DutyInfo struct {
		ID        int    `db:"id" json:"id"`
		Cname     string `db:"cname" json:"cname,omitempty"`
		UserIDs   string `db:"user_ids" json:"user_ids,omitempty"`
		BeginTime int    `db:"begin_time" json:"begin_time,omitempty"`
		EndTime   int    `db:"end_time" json:"end_time,omitempty"`
	}
	// DutyStatus is alarm status for users in the duty
	DutyStatus struct {
		DutyID       int    `db:"duty_id" json:"duty_id"`
		AlarmTeamID  int    `db:"alarm_team_id" json:"alarm_team_id,omitempty"`
		StartTime    int    `db:"start_time" json:"start_time,omitempty"`
		EndTime      int    `db:"end_time" json:"end_time,omitempty"`
		EmailStatus  int    `db:"email_status" json:"email,omitempty"`
		PhoneStatus  int    `db:"phone_status" json:"phone,omitempty"`
		WechatStatus int    `db:"wechat_status" json:"wechat,omitempty"`
		SmsStatus    int    `db:"sms_status" json:"sms,omitempty"`
		Step         int    `db:"step" json:"step,omitempty"`
		UserName     string `db:"-" json:"user_name,omitempty"`
		UserNameCN   string `db:"-" json:"user_name_cn,omitempty"`
	}
)

// IsInTime checks whether the time is in range for the duty to alarm
func (duty *DutyInfo) IsInTime(t int64) bool {
	if duty.BeginTime == duty.EndTime {
		return false
	}
	now := time.Unix(t, 0).Format("15:04")
	t1 := time.Unix(int64(duty.BeginTime), 0).In(time.UTC).Format("15:04:05")
	t2 := time.Unix(int64(duty.EndTime), 0).In(time.UTC).Format("15:04:05")
	if duty.EndTime > duty.BeginTime {
		return now >= t1 && now <= t2
	}
	return !(now >= t2 && now <= t1)
}

// AlarmAction is action binding to template to alarm users
type AlarmAction struct {
	ID                 int    `json:"id" db:"id"`
	Uic                string `json:"uic" db:"uic"`
	URL                string `json:"url,omitempty" db:"url"`
	Callback           bool   `json:"callback,omitempty" db:"callback"`
	BeforeCallbackSms  bool   `json:"before_callback_sms,omitempty" db:"before_callback_sms"`
	BeforeCallbackMail bool   `json:"before_callback_mail,omitempty" db:"before_callback_mail"`
	AfterCallbackSms   bool   `json:"after_callback_sms,omitempty" db:"after_callback_sms"`
	AfterCallbackMail  bool   `json:"after_callback_mail,omitempty" db:"after_callback_mail"`
}

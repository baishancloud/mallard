package sqldata

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

// Alarms maintains all alarming data
type Alarms struct {
	ForStrategies   map[int]*AlarmsForStrategy    `json:"for_strategies,omitempty"`
	Actions         map[int]*models.AlarmAction   `json:"actions,omitempty"`
	UsersInfos      map[int]*models.UserInfo      `json:"users_infos,omitempty"`
	TeamInfos       map[int]*models.TeamInfo      `json:"team_infos,omitempty"`
	OuterInfos      map[int]*models.OuterUserInfo `json:"outer_infos,omitempty"`
	DutyInfos       map[int]*models.DutyInfo      `json:"duty_infos,omitempty"`
	StrategyActions map[int]*models.AlarmAction   `json:"strategy_actions,omitempty"`
	Templates       map[int]*models.Template      `json:"templates,omitempty"`
	CRC             uint32                        `json:"crc,omitempty"`

	cachedRequests map[int]map[string]*AlarmSendRequest
	cachedLock     sync.RWMutex
}

// AlarmsForStrategy is alarms info for one strategy
type AlarmsForStrategy struct {
	UserStatus      []*models.UserStatus     `json:"user_status,omitempty"`
	TeamStatus      []*models.TeamUserStatus `json:"team_status,omitempty"`
	DutyStatus      []*models.DutyStatus     `json:"duty_status,omitempty"`
	OuterStatus     []*models.UserStatus     `json:"outer_status,omitempty"`
	ActionUic       string                   `json:"action_uic,omitempty"`
	TemplateName    string                   `json:"template_name,omitempty"`
	MaxStep         int                      `json:"max_step,omitempty"`
	MaxStepDuration int                      `json:"max_step_duration,omitempty"`
}

func (da *Data) buildAlarms() {
	alarms := &Alarms{
		Actions:         da.AlarmActions,
		ForStrategies:   make(map[int]*AlarmsForStrategy, len(da.Strategies)),
		UsersInfos:      da.UserInfos,
		TeamInfos:       da.TeamInfos,
		OuterInfos:      da.OuterUserInfos,
		DutyInfos:       da.DutyInfos,
		StrategyActions: make(map[int]*models.AlarmAction, len(da.Strategies)),
	}
	for sid, st := range da.Strategies {
		forSt := &AlarmsForStrategy{
			ActionUic:       "000",
			MaxStep:         st.MaxStep,
			MaxStepDuration: st.Step,
		}
		// fill uic
		tpl := da.Templates[st.TemplateID]
		if tpl != nil {
			forSt.TemplateName = tpl.Name
			action := da.AlarmActions[tpl.ActionID]
			if action != nil {
				forSt.ActionUic = action.Uic
				alarms.StrategyActions[sid] = action
			}
		}
		// fill all status
		users := da.AlarmUsersStatus[sid]
		if len(users) > 0 {
			forSt.UserStatus = users
		}
		outers := da.OuterUsersStatus[sid]
		if len(outers) > 0 {
			forSt.OuterStatus = outers
		}
		teams := da.TeamStrategies[sid]
		for _, team := range teams {
			users := da.TeamUsersStatus[team.TeamID]
			if len(users) > 0 {
				forSt.TeamStatus = append(forSt.TeamStatus, users...)
			}
			dutys := da.DutyUsersStatus[team.TeamID]
			if len(dutys) > 0 {
				forSt.DutyStatus = append(forSt.DutyStatus, dutys...)
			}
		}
		alarms.ForStrategies[sid] = forSt
	}
	b, _ := json.Marshal(alarms)
	alarms.CRC = crc32.ChecksumIEEE(b)
	da.alarms = alarms
	log.Debug("alarms-for-strategy", "strategies", len(alarms.ForStrategies))
}

// BuildAll builds all data
func (al *Alarms) BuildAll() {
	for sid, forSt := range al.ForStrategies {
		userKeys := make(AlarmMethods, len(forSt.TeamStatus))
		// fill user status
		for _, us := range forSt.UserStatus {
			userName := al.getUserName(us.UserID)
			if userName != "" {
				key := fmt.Sprintf("user-%d", us.Step)
				userKeys[key] = append(userKeys[key], AlarmUserMethod{
					UserName: userName,
					Email:    us.EmailStatus,
					Wechat:   us.WechatStatus,
					SMS:      us.SmsStatus,
					Phone:    us.PhoneStatus,
					Step:     us.Step,
					UserID:   us.UserID,
				})
			}
		}
		// fill team user status
		for _, us := range forSt.TeamStatus {
			userName := al.getUserName(us.UserID)
			if userName == "" {
				continue
			}
			team := al.TeamInfos[us.AlarmTeamID]
			if team == nil {
				continue
			}
			key := fmt.Sprintf("%s-%d", team.Name, us.Step)
			userKeys[key] = append(userKeys[key], AlarmUserMethod{
				UserName: userName,
				Email:    us.EmailStatus,
				Wechat:   us.WechatStatus,
				SMS:      us.SmsStatus,
				Phone:    us.PhoneStatus,
				Step:     us.Step,
				TeamID:   us.AlarmTeamID,
				UserID:   us.UserID,
			})
		}
		for _, us := range forSt.DutyStatus {
			duty := al.DutyInfos[us.DutyID]
			if duty == nil {
				continue
			}
			team := al.TeamInfos[us.AlarmTeamID]
			if team == nil {
				continue
			}
			uIDS := strings.Split(duty.UserIDs, ",")
			if len(uIDS) == 0 {
				continue
			}
			for _, uidStr := range uIDS {
				uid, _ := strconv.Atoi(uidStr)
				if uid == 0 {
					continue
				}
				username := al.getUserName(uid)
				if username == "" {
					continue
				}
				key := fmt.Sprintf("duty-%d-%d", duty.ID, us.Step)
				userKeys[key] = append(userKeys[key], AlarmUserMethod{
					UserName: username,
					Email:    us.EmailStatus,
					Wechat:   us.WechatStatus,
					SMS:      us.SmsStatus,
					Phone:    us.PhoneStatus,
					Step:     us.Step,
					DutyID:   us.DutyID,
					TeamID:   us.AlarmTeamID,
					UserID:   uid,
				})
			}
		}
		reqs := al.MethodsMap(sid, userKeys)
		if len(reqs) == 0 {
			continue
		}
		al.cachedLock.Lock()
		if al.cachedRequests == nil {
			al.cachedRequests = make(map[int]map[string]*AlarmSendRequest, len(al.ForStrategies))
		}
		if forSt.MaxStep > 0 && forSt.MaxStepDuration > 0 {
			for i := 1; i < forSt.MaxStep; i++ {
				realStep := i * forSt.MaxStepDuration
				for key, req := range reqs {
					if req.Step == 0 {
						key2 := fmt.Sprintf("%s-maxstep-%d", key, realStep)
						cp := new(AlarmSendRequest)
						*cp = *req
						cp.Step = realStep
						reqs[key2] = cp
					}
				}
			}
		}
		for _, req := range reqs {
			req.Uic = forSt.ActionUic
			req.Template = forSt.TemplateName
		}
		al.cachedRequests[sid] = reqs
		al.cachedLock.Unlock()
	}
}

func (al *Alarms) getUserName(uid int) string {
	user := al.UsersInfos[uid]
	if user == nil {
		return ""
	}
	return user.Name
}

type (
	// AlarmUserMethod is method list of one user's alarm
	AlarmUserMethod struct {
		UserName string `json:"user_name,omitempty"`
		Email    int    `json:"email,omitempty"`
		Wechat   int    `json:"wechat,omitempty"`
		SMS      int    `json:"sms,omitempty"`
		Phone    int    `json:"phone,omitempty"`
		Step     int    `json:"step,omitempty"`
		UserID   int    `json:"user_id,omitempty"`
		TeamID   int    `json:"team_id,omitempty"`
		DutyID   int    `json:"duty_id,omitempty"`
	}
	// AlarmMethods is groups of user alarms by same team key
	AlarmMethods map[string][]AlarmUserMethod
	// AlarmSendRequest is alarm request data for one strategy and one team key
	AlarmSendRequest struct {
		Emails   []string `json:"emails,omitempty"`
		Wechats  []string `json:"wechats,omitempty"`
		SMSs     []string `json:"smss,omitempty"`
		Phones   []string `json:"phones,omitempty"`
		Team     string   `json:"team,omitempty"`
		TeamCN   string   `json:"team_cn,omitempty"`
		Step     int      `json:"step,omitempty"`
		Uic      string   `json:"uic,omitempty"`
		Template string   `json:"template,omitempty"`
		Keys     []string `json:"keys,omitempty"`
	}
)

func (asr AlarmSendRequest) Unique() string {
	s := asr.Emails
	s = append(s, asr.Wechats...)
	s = append(s, asr.SMSs...)
	s = append(s, asr.Phones...)
	return strings.Join(s, "-")
}

// Line prints send-request to line
func (asr AlarmSendRequest) Line() string {
	s := make([]string, 0, 6)
	if len(asr.Emails) == 0 {
		s = append(s, ",")
	} else {
		s = append(s, strings.Join(asr.Emails, ","))
	}
	if len(asr.Wechats) == 0 {
		s = append(s, ",")
	} else {
		s = append(s, strings.Join(asr.Wechats, ","))
	}
	if len(asr.SMSs) == 0 {
		s = append(s, ",")
	} else {
		s = append(s, strings.Join(asr.SMSs, ","))
	}

	if len(asr.Phones) == 0 {
		s = append(s, ",")
	} else {
		s = append(s, strings.Join(asr.Phones, ","))
	}
	teamName := asr.Template + ","
	if asr.Template != "000" {
	}
	if asr.Team == "" {
		teamName += "000"
	} else {
		teamName += asr.Team
	}
	teamName += ","
	if asr.TeamCN == "" {
		teamName += "000"
	} else {
		teamName += asr.TeamCN
	}
	s = append(s, teamName)
	return strings.Join(s, "|")
}

type tempMethod struct {
	Key    string
	Method AlarmUserMethod
	Step   int
}

// MethodsMap converts methods to requests
func (al *Alarms) MethodsMap(sid int, methods AlarmMethods) map[string]*AlarmSendRequest {
	userKeysCache := make(map[int][]tempMethod, len(methods))
	for key, users := range methods {
		for _, user := range users {
			userKeysCache[user.UserID] = append(userKeysCache[user.UserID], tempMethod{
				Key:    key,
				Method: user,
				Step:   user.Step,
			})
		}
	}
	reqs := make(map[string]*AlarmSendRequest)
	for uid, users := range userKeysCache {
		if len(users) > 1 {
			for _, tmpUser := range users {
				key := fmt.Sprintf("user-merge-%d-%d", uid, tmpUser.Step)
				req := reqs[key]
				if req == nil {
					req = &AlarmSendRequest{
						Step: tmpUser.Step,
					}
					reqs[key] = req
				}
				user := tmpUser.Method
				req.Team, req.TeamCN = al.getTeamNames(user)
				if user.Email > 0 {
					req.Emails = append(req.Emails, user.UserName)
				}
				if user.Wechat > 0 {
					req.Wechats = append(req.Wechats, user.UserName)
				}
				if user.SMS > 0 {
					req.SMSs = append(req.SMSs, user.UserName)
				}
				if user.Phone > 0 {
					req.Phones = append(req.Phones, user.UserName)
				}
				req.Keys = append(req.Keys, tmpUser.Key)
			}
			continue
		}
		user := users[0].Method
		req := reqs[users[0].Key]
		if req == nil {
			req = &AlarmSendRequest{
				Step: user.Step,
			}
			reqs[users[0].Key] = req
		}
		req.Team, req.TeamCN = al.getTeamNames(user)
		if user.Email > 0 {
			req.Emails = append(req.Emails, user.UserName)
		}
		if user.Wechat > 0 {
			req.Wechats = append(req.Wechats, user.UserName)
		}
		if user.SMS > 0 {
			req.SMSs = append(req.SMSs, user.UserName)
		}
		if user.Phone > 0 {
			req.Phones = append(req.Phones, user.UserName)
		}
	}
	for _, req := range reqs {
		req.Emails = utils.StringSliceUnique(req.Emails)
		sort.Sort(sort.StringSlice(req.Emails))

		req.Wechats = utils.StringSliceUnique(req.Wechats)
		sort.Sort(sort.StringSlice(req.Wechats))

		req.SMSs = utils.StringSliceUnique(req.SMSs)
		sort.Sort(sort.StringSlice(req.SMSs))

		req.Phones = utils.StringSliceUnique(req.Phones)
		sort.Sort(sort.StringSlice(req.Phones))

		if len(req.Keys) > 0 {
			sort.Sort(sort.StringSlice(req.Keys))
		}
	}
	return reqs
}

func (al *Alarms) getTeamNames(method AlarmUserMethod) (string, string) {
	if method.TeamID > 0 {
		if team := al.TeamInfos[method.TeamID]; team != nil {
			return team.Name, team.Title
		}
	}
	if method.DutyID > 0 {
		if duty := al.DutyInfos[method.DutyID]; duty != nil {
			return fmt.Sprintf("Duty-%d", duty.ID), duty.Cname
		}
	}
	return "000", "000"
}

// ForStrategy returns alarm requests for one strategy
func (al *Alarms) ForStrategy(sid int) map[string]*AlarmSendRequest {
	al.cachedLock.RLock()
	defer al.cachedLock.RUnlock()
	return al.cachedRequests[sid]
}

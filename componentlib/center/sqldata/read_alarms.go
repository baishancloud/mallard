package sqldata

import (
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	selectAlarmUserInfoSQL = "SELECT id,name,cnname,email,phone,im,qq,role FROM user WHERE status > 0 ORDER BY id ASC"
)

// ReadUserInfo reads alarm users
func ReadUserInfo() (map[int]*models.UserInfo, error) {
	rows, err := uicDB.Queryx(selectAlarmUserInfoSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	m := make(map[int]*models.UserInfo)
	for rows.Next() {
		u := new(models.UserInfo)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.ID > 0 {
			m[u.ID] = u
			count++
		}
	}
	return m, nil
}

var (
	selectTeamInfoSQL = "SELECT id,name,title,remark FROM alarm_team"
)

// ReadTeamInfo reads teams info
func ReadTeamInfo() (map[int]*models.TeamInfo, error) {
	rows, err := uicDB.Queryx(selectTeamInfoSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int]*models.TeamInfo)
	var count int
	for rows.Next() {
		u := new(models.TeamInfo)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.ID > 0 {
			m[u.ID] = u
			count++
		}
	}
	return m, nil
}

var (
	selectDutyInfoSQL = "SELECT id,cname,user_ids,begin_time,end_time FROM duty_team"
)

// ReadDutyInfo query duty info
func ReadDutyInfo() (map[int]*models.DutyInfo, error) {
	rows, err := uicDB.Queryx(selectDutyInfoSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int]*models.DutyInfo)
	var count int
	var now = time.Now().Unix()
	// var nowStr = time.Now().Format("15:04:05")
	for rows.Next() {
		u := new(models.DutyInfo)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.ID > 0 {
			if u.UserIDs == "" {
				//c.logger.Debug("skip-duty-no-users", "du", u.Cname)
				continue
			}
			if u.BeginTime == u.EndTime || !u.IsInTime(now) {
				/*t1 := time.Unix(int64(u.BeginTime), 0).In(time.UTC).Format("15:04:05")
				t2 := time.Unix(int64(u.EndTime), 0).In(time.UTC).Format("15:04:05")
				c.logger.Debug("skip-duty-no-intime", "du", u.Cname, "range", t1+" ~ "+t2, "now", nowStr)*/
				continue
			}
			m[u.ID] = u
			count++
		}
	}
	return m, nil
}

var (
	selectOuterUserSQL = "SELECT u.id AS id ,u.name AS name,u.email AS email,u.phone AS phone,u.company_id AS company_id ,c.name AS company_name FROM outer_user AS u JOIN outer_company AS c ON c.id = u.company_id"
)

// ReadOuterUsers reads outer users
func ReadOuterUsers() (map[int]*models.OuterUserInfo, error) {
	rows, err := uicDB.Queryx(selectOuterUserSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	m := make(map[int]*models.OuterUserInfo)
	for rows.Next() {
		u := new(models.OuterUserInfo)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}

		if u.ID > 0 {
			m[u.ID] = u
			count++
		}
	}
	return m, nil
}

var (
	selectTeamUsersSQL = "SELECT user_id,alarm_team_id,start_time,end_time,email_status,phone_status,wechat_status,sms_status,step FROM alarm_team_user"
)

// ReadTeamUsers reads users status in team
func ReadTeamUsers() (map[int][]*models.TeamUserStatus, error) {
	rows, err := uicDB.Queryx(selectTeamUsersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	var now = time.Now().Unix()
	m := make(map[int][]*models.TeamUserStatus)
	for rows.Next() {
		u := new(models.TeamUserStatus)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.UserID > 0 && u.AlarmTeamID > 0 {
			if !u.IsInTime(now) {
				/*t1 := time.Unix(int64(u.StartTime), 0).In(time.UTC).Format("15:04:05")
				t2 := time.Unix(int64(u.EndTime), 0).In(time.UTC).Format("15:04:05")
				c.logger.Debug("skip-team-user", "team", u.AlarmTeamID, "user", u.UserID, "range", t1+" ~ "+t2)*/
				continue
			}
			m[u.AlarmTeamID] = append(m[u.AlarmTeamID], u)
			count++
		}
	}
	return m, nil
}

var (
	selectAlarmActionsSQL = "SELECT id,uic,url,callback,before_callback_sms,before_callback_mail,after_callback_sms,after_callback_mail FROM action"
)

// ReadAlarmActions query alarm actions as id-action map
func ReadAlarmActions() (map[int]*models.AlarmAction, error) {
	rows, err := portalDB.Queryx(selectAlarmActionsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var (
		count   int
		actions = make(map[int]*models.AlarmAction)
	)
	for rows.Next() {
		var act models.AlarmAction
		if err = rows.StructScan(&act); err != nil {
			continue
		}
		if act.ID > 0 {
			actions[act.ID] = &act
		}
		count++
	}
	return actions, nil
}

var (
	selectUserStatusSQL = "SELECT user_id,strategy_id,email_status,phone_status,wechat_status,sms_status,step FROM alarm_strategy_user"
)

// ReadAlarmUsers return users alarm status
func ReadAlarmUsers() (map[int][]*models.UserStatus, error) {
	rows, err := portalDB.Queryx(selectUserStatusSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	m := make(map[int][]*models.UserStatus)
	for rows.Next() {
		u := new(models.UserStatus)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.UserID > 0 && u.StrategyID > 0 && u.IsEnable() {
			m[u.StrategyID] = append(m[u.StrategyID], u)
			count++
		}
	}
	return m, nil
}

var (
	selectTeamStatusSQL = "SELECT team_id,strategy_id,step FROM alarm_strategy_team"
)

// ReadTeamStrategies reads alarm teams relation to strategy
func ReadTeamStrategies() (map[int][]*models.TeamStrategy, error) {
	rows, err := portalDB.Queryx(selectTeamStatusSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	m := make(map[int][]*models.TeamStrategy)
	for rows.Next() {
		u := new(models.TeamStrategy)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.TeamID > 0 && u.StrategyID > 0 {
			m[u.StrategyID] = append(m[u.StrategyID], u)
			count++
		}
	}
	return m, nil
}

var (
	selectCompanyUserStatusSQL = "SELECT user_id,strategy_id,email_status,phone_status,wechat_status,sms_status,step FROM alarm_strategy_company_user"
)

// ReadAlarmOuter reads outer users to strategy
func ReadAlarmOuter() (map[int][]*models.UserStatus, error) {
	rows, err := portalDB.Queryx(selectCompanyUserStatusSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	m := make(map[int][]*models.UserStatus)
	for rows.Next() {
		u := new(models.UserStatus)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.UserID > 0 && u.StrategyID > 0 && u.IsEnable() {
			m[u.StrategyID] = append(m[u.StrategyID], u)
			count++
		}
	}
	return m, nil
}

var (
	selectA2DutyUsersSQL = "SELECT duty_id,alarm_team_id,start_time,end_time,email_status,phone_status,wechat_status,sms_status,step FROM alarm_team_duty"
)

// ReadAlarmDuty reads duty data to each strategy
func ReadAlarmDuty() (map[int][]*models.DutyStatus, error) {
	rows, err := uicDB.Queryx(selectA2DutyUsersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	m := make(map[int][]*models.DutyStatus)
	for rows.Next() {
		u := new(models.DutyStatus)
		if err = rows.StructScan(u); err != nil {
			return nil, err
		}
		if u.DutyID > 0 && u.AlarmTeamID > 0 {
			m[u.AlarmTeamID] = append(m[u.AlarmTeamID], u)
			count++
		}
	}
	return m, nil
}

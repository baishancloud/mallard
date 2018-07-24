package sqldata

import (
	"time"

	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/jmoiron/sqlx"
)

var (
	portalDB *sqlx.DB
	uicDB    *sqlx.DB

	log = zaplog.Zap("sqldata")
)

// SetDB sets db object
func SetDB(portal, uic *sqlx.DB) {
	portalDB = portal
	uicDB = uic
}

func checkDB(db *sqlx.DB) {
	if db == nil {
		log.Fatal("db-nil")
	}
}

var (
	cachedData = new(Data)
)

// Read reads database to generate Data object
func Read() (*Data, error) {
	data := new(Data)
	var err error

	if data.Templates, err = ReadTemplates(); err != nil {
		return nil, err
	}
	log.Debug("read-templates", "templates", len(data.Templates))

	if data.Strategies, err = ReadStrategies(); err != nil {
		return nil, err
	}
	log.Debug("read-strategies", "strategies", len(data.Strategies))

	if data.Expressions, err = ReadExpressions(); err != nil {
		return nil, err
	}
	log.Debug("read-expressions", "expressions", len(data.Expressions))

	if data.GroupPlugins, err = ReadGroupPlugins(); err != nil {
		return nil, err
	}
	log.Debug("read-group-plugins", "groups", len(data.GroupPlugins))

	if data.GroupHosts, err = ReadGroupHosts(); err != nil {
		return nil, err
	}
	log.Debug("read-group-hosts", "hosts", len(data.GroupHosts))

	if data.GroupTemplates, err = ReadGroupTemplates(); err != nil {
		return nil, err
	}
	log.Debug("read-group-templates", "groups", len(data.GroupTemplates))

	if data.GroupNames, err = ReadGroupNames(); err != nil {
		return nil, err
	}
	log.Debug("read-group-names", "groups", len(data.GroupNames))

	if data.HostNames, data.HostInfos, data.HostMaintains, data.HostLiveInfos, err = ReadHosts(); err != nil {
		return nil, err
	}
	log.Debug("read-hosts", "names", len(data.HostNames), "infos", len(data.HostInfos), "maintains", len(data.HostMaintains))

	if data.UserInfos, err = ReadUserInfo(); err != nil {
		return nil, err
	}
	log.Debug("read-users", "users", len(data.UserInfos))

	if data.TeamInfos, err = ReadTeamInfo(); err != nil {
		return nil, err
	}
	log.Debug("read-teams", "teams", len(data.TeamInfos))

	if data.DutyInfos, err = ReadDutyInfo(); err != nil {
		return nil, err
	}
	log.Debug("read-dutys", "dutys", len(data.DutyInfos))

	if data.OuterUserInfos, err = ReadOuterUsers(); err != nil {
		return nil, err
	}
	log.Debug("read-outer-users", "users", len(data.OuterUserInfos))

	if data.TeamUsersStatus, err = ReadTeamUsers(); err != nil {
		return nil, err
	}
	log.Debug("read-team-users", "teams", len(data.TeamUsersStatus))

	if data.AlarmActions, err = ReadAlarmActions(); err != nil {
		return nil, err
	}
	log.Debug("read-alarm-actions", "actions", len(data.AlarmActions))

	if data.AlarmUsersStatus, err = ReadAlarmUsers(); err != nil {
		return nil, err
	}
	log.Debug("read-users-status", "users", len(data.AlarmUsersStatus))

	if data.TeamStrategies, err = ReadTeamStrategies(); err != nil {
		return nil, err
	}
	log.Debug("read-team-strategies", "strategies", len(data.TeamStrategies))

	if data.OuterUsersStatus, err = ReadAlarmOuter(); err != nil {
		return nil, err
	}
	log.Debug("read-outers-status", "outers", len(data.OuterUsersStatus))

	if data.DutyUsersStatus, err = ReadAlarmDuty(); err != nil {
		return nil, err
	}
	log.Debug("read-duty-status", "dutys", len(data.DutyUsersStatus))

	return data, nil
}

// Sync starts to updating cached data from database
// when update, send to channel
func Sync(interval time.Duration, ch chan<- *Data) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		st := time.Now()
		data, err := Read()
		if err != nil {
			log.Error("read-error", "error", err)
			buildFailCount.Incr(1)
		} else {
			oldHash := cachedData.Hash()
			hash := data.GenerateHash()
			if oldHash != hash {
				data.build()
				cachedData = data
				if ch != nil {
					ch <- data
				}
				ms := utils.DurationMS(time.Since(st))
				log.Info("update", "hash", hash, "old", oldHash, "ms", ms)
				buildDurationAvg.Set(ms)
				buildChangeCount.Incr(1)
			} else {
				ms := utils.DurationMS(time.Since(st))
				log.Info("keep", "hash", oldHash, "ms", ms)
				buildDurationAvg.Set(ms)
			}
		}
		<-ticker.C
	}
}

package sqldata

import (
	"encoding/json"
	"strconv"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

// Data is all data to maintains sqldata
type Data struct {
	Strategies   map[int]*models.Strategy    `json:"strategies,omitempty"`
	Expressions  map[int]*models.Expression  `json:"expressions,omitempty"`
	Templates    map[int]*models.Template    `json:"templates,omitempty"`
	AlarmActions map[int]*models.AlarmAction `json:"alarm_actions,omitempty"`

	GroupPlugins   map[int][]string `json:"group_plugins,omitempty"`
	GroupHosts     map[int][]int    `json:"group_hosts,omitempty"`
	GroupTemplates map[int][]int    `json:"group_templates,omitempty"`
	GroupNames     map[int]string   `json:"group_names,omitempty"`

	HostNames     map[int]string                 `json:"host_names,omitempty"`
	HostInfos     map[string]string              `json:"-"`
	HostLiveInfos map[string][]interface{}       `json:"-"`
	HostMaintains map[string]int64               `json:"host_maintains,omitempty"`
	HostServices  map[string]*models.HostService `json:"-"`

	UserInfos      map[int]*models.UserInfo      `json:"user_infos,omitempty"`
	TeamInfos      map[int]*models.TeamInfo      `json:"team_infos,omitempty"`
	DutyInfos      map[int]*models.DutyInfo      `json:"duty_infos,omitempty"`
	OuterUserInfos map[int]*models.OuterUserInfo `json:"outer_user_infos,omitempty"`

	TeamUsersStatus  map[int][]*models.TeamUserStatus `json:"team_users_status,omitempty"`
	AlarmUsersStatus map[int][]*models.UserStatus     `json:"alarm_users_status,omitempty"`
	OuterUsersStatus map[int][]*models.UserStatus     `json:"outer_users_status,omitempty"`
	DutyUsersStatus  map[int][]*models.DutyStatus     `json:"duty_users_status,omitempty"`
	TeamStrategies   map[int][]*models.TeamStrategy   `json:"team_strategies,omitempty"`

	endpoints *Endpoints
	alarms    *Alarms
	hash      string
}

// GenerateHash generates hash of all data
func (da *Data) GenerateHash() string {
	b, _ := json.Marshal(da)
	hash := utils.MD5HashBytes(b)
	da.hash = hash
	return hash
}

// Hash gets hash after generating
func (da *Data) Hash() string {
	return da.hash
}

var (
	buildDurationAvg    = expvar.NewAverage("cache.build_duration", 10)
	buildFailCount      = expvar.NewDiff("cache.build_fail")
	buildChangeCount    = expvar.NewDiff("cache.build_change")
	alarmMessagesCount  = expvar.NewDiff("cache.alarm_messages")
	endpointsCacheCount = expvar.NewDiff("cache.epconfigs")
	groupsCount         = expvar.NewDiff("cache.groups")
	hostsCount          = expvar.NewDiff("cache.hosts")
	strategiesCount     = expvar.NewDiff("cache.strategies")
	templatesCount      = expvar.NewDiff("cache.templates")
)

// InitExpvars inits expvars mannually
func InitExpvars() {
	expvar.Register(alarmMessagesCount, buildDurationAvg, buildChangeCount,
		buildFailCount,
		endpointsCacheCount, groupsCount, hostsCount,
		strategiesCount, templatesCount)
}

func (da *Data) build() {
	groupsCount.Set(int64(len(da.GroupNames)))
	hostsCount.Set(int64(len(da.HostNames)))
	strategiesCount.Set(int64(len(da.Strategies)))
	templatesCount.Set(int64(len(da.Templates)))

	da.buildEndpoints()
	if da.endpoints != nil {
		da.endpoints.BuildAll()
		endpointsCacheCount.Set(int64(len(da.endpoints.cachedConfigs)))
	}

	da.buildAlarms()
	if da.alarms != nil {
		da.alarms.BuildAll()
		alarmMessagesCount.Set(int64(len(da.alarms.cachedRequests)))
	}
}

// EndpointsAll returns all endpoints data
func EndpointsAll() *Endpoints {
	if cachedData != nil {
		return cachedData.endpoints
	}
	return nil
}

// EndpointOne returns one endpoint config
func EndpointOne(endpoint string) *models.EndpointConfig {
	if cachedData != nil {
		return cachedData.endpoints.Endpoint(endpoint)
	}
	return nil
}

// AlarmsAll returns all alarms data
func AlarmsAll() *Alarms {
	if cachedData != nil {
		return cachedData.alarms
	}
	return nil
}

// AlarmsRequests returns all alarm requests data
func AlarmsRequests() map[int]map[string]*AlarmSendRequest {
	if cachedData != nil {
		cachedData.alarms.cachedLock.RLock()
		defer cachedData.alarms.cachedLock.RUnlock()
		return cachedData.alarms.cachedRequests
	}
	return nil
}

// AlarmsForOneStrategy is alarm request for one strategy
func AlarmsForOneStrategy(sid int) map[string]*AlarmSendRequest {
	if cachedData != nil {
		return cachedData.alarms.ForStrategy(sid)
	}
	return nil
}

// AlarmTeamBy gets alarm team by name or id
func AlarmTeamBy(by string, value string) (*models.TeamInfo, []*models.TeamUserStatus) {
	if cachedData == nil {
		return nil, nil
	}
	var team *models.TeamInfo
	if by == "name" {
		for _, t := range cachedData.TeamInfos {
			if t.Name == value {
				team = t
				break
			}
		}
	} else if by == "id" {
		idInt, _ := strconv.Atoi(value)
		team = cachedData.TeamInfos[idInt]
	}
	if team == nil {
		return nil, nil
	}
	status := cachedData.TeamUsersStatus[team.ID]
	for _, u := range status {
		user := cachedData.UserInfos[u.UserID]
		if user != nil {
			u.UserName = user.Name
			u.UserNameCN = user.Cnname
		}
	}
	return team, status
}

// StrategiesAll gets all strategies
func StrategiesAll() map[int]*models.Strategy {
	if cachedData == nil {
		return nil
	}
	return cachedData.Strategies
}

// ExpressionsAll gets all expressions
func ExpressionsAll() map[int]*models.Expression {
	if cachedData == nil {
		return nil
	}
	return cachedData.Expressions
}

// TemplatesAll gets all templates
func TemplatesAll() map[int]*models.Template {
	if cachedData == nil {
		return nil
	}
	return cachedData.Templates
}

// GroupPluginsAll gets all group plugins
func GroupPluginsAll() map[int][]string {
	if cachedData == nil {
		return nil
	}
	return cachedData.GroupPlugins
}

// DataHash is hash of all data
func DataHash() string {
	if cachedData == nil {
		return ""
	}
	return cachedData.hash
}

// HostInfosAll returns all hosts info, maintains and live-at times
func HostInfosAll() map[string][]interface{} {
	if cachedData == nil {
		return nil
	}
	return cachedData.HostLiveInfos
}

// HostServices returns host services
func HostServices() map[string]*models.HostService {
	if cachedData == nil {
		return nil
	}
	return cachedData.HostServices
}

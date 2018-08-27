package sqldata

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

// Len return length of cache endpoint config
func (eps *Endpoints) Len() int {
	eps.cachedLock.RLock()
	defer eps.cachedLock.RUnlock()
	return len(eps.cachedConfigs)
}

// Endpoints is all configs to build each config of endpoints
type Endpoints struct {
	GroupStrategy map[string][]int
	HostGroupKeys map[string]string
	GroupPlugins  map[string][]string
	Strategies    map[int]*models.Strategy
	HostMaintians map[string]int64
	HostConfigs   map[string]*models.HostConfig
	CRC           uint32

	cachedConfigs map[string]*models.EndpointConfig
	cachedLock    sync.RWMutex
}

// Endpoint generate one endpoint config
func (eps *Endpoints) Endpoint(endpoint string) *models.EndpointConfig {
	eps.cachedLock.RLock()
	if len(eps.cachedConfigs) == 0 {
		eps.cachedLock.RUnlock()
		return nil
	}
	key := eps.HostGroupKeys[endpoint]
	epData := eps.cachedConfigs[key]
	eps.cachedLock.RUnlock()
	if epData != nil {
		return epData
	}
	slist := eps.GroupStrategy[key]
	epData = &models.EndpointConfig{
		Plugins: eps.GroupPlugins[key],
		Builtin: &models.EndpointBuiltin{},
	}
	if eps.HostMaintians[endpoint] > 0 {
		eps.cachedLock.Lock()
		eps.HostGroupKeys[endpoint] = endpoint
		eps.cachedConfigs[endpoint] = epData
		eps.cachedLock.Unlock()
	} else if agentCfg := eps.HostConfigs[endpoint]; agentCfg != nil && agentCfg.AgentStatus == models.AgentStatusIgnore {
		eps.cachedLock.Lock()
		epData.Plugins = []string{"null"}
		eps.HostGroupKeys[endpoint] = endpoint
		eps.cachedConfigs[endpoint] = epData
		eps.cachedLock.Unlock()
	} else {
		for _, sid := range slist {
			s := eps.Strategies[sid]
			if s != nil {
				epData.Strategies = append(epData.Strategies, s)
			}
		}
		for _, s := range epData.Strategies {
			setBuiltin(epData, s)
		}
	}
	eps.cachedLock.Lock()
	eps.cachedConfigs[key] = epData
	eps.cachedLock.Unlock()
	return epData
}

func setBuiltin(ca *models.EndpointConfig, s *models.Strategy) {
	if s.Metric == "net_port_listen" {
		tags, err := models.ExtractTags(s.TagString)
		if err != nil {
			return
		}
		if port, ok := tags["port"]; ok && port != "" {
			portInt, _ := strconv.ParseInt(port, 10, 64)
			if portInt > 0 {
				ca.Builtin.Ports = append(ca.Builtin.Ports, portInt)
			}
		}
	}
}

// BuildAll generetes all raw config data for host-group-keys
func (eps *Endpoints) BuildAll() {
	eps.cachedLock.Lock()
	defer eps.cachedLock.Unlock()
	if eps.cachedConfigs == nil {
		eps.cachedConfigs = make(map[string]*models.EndpointConfig)
	}
	for key, slist := range eps.GroupStrategy {
		if len(slist) < 1 {
			continue
		}
		epData := &models.EndpointConfig{
			Plugins: eps.GroupPlugins[key],
			Builtin: &models.EndpointBuiltin{},
		}
		for _, sid := range slist {
			s := eps.Strategies[sid]
			if s != nil {
				epData.Strategies = append(epData.Strategies, s.ToSimple())
			}
		}
		for _, s := range epData.Strategies {
			setBuiltin(epData, s)
		}
		eps.cachedConfigs[key] = epData
	}
	// set maintains agent
	for endpoint := range eps.HostMaintians {
		key := eps.HostGroupKeys[endpoint]
		cacheData := eps.cachedConfigs[key]
		epData := &models.EndpointConfig{
			Builtin: &models.EndpointBuiltin{},
		}
		if cacheData != nil {
			epData.Plugins = cacheData.Plugins
		}
		eps.HostGroupKeys[endpoint] = endpoint
		eps.cachedConfigs[endpoint] = epData
	}
	// set ignored agent
	for endpoint, agentCfg := range eps.HostConfigs {
		if agentCfg.AgentStatus == models.AgentStatusIgnore {
			epData := &models.EndpointConfig{
				Builtin: &models.EndpointBuiltin{},
				Plugins: []string{"null"},
			}
			eps.HostGroupKeys[endpoint] = endpoint
			eps.cachedConfigs[endpoint] = epData
		}
	}
}

func (da *Data) buildEndpoints() {
	eps := &Endpoints{}
	var skipCount int
	realStrategies := make(map[int]*models.Strategy, len(da.Strategies))
	for k, s := range da.Strategies {
		if s.IsEnable() {
			realStrategies[k] = s
			continue
		}
		skipCount++
	}
	log.Debug("skip-strategies", "count", skipCount)

	// template:1 -- n:strategy
	templateStrategies := make(map[int][]int, len(da.Templates))
	for _, st := range realStrategies {
		tpl := da.Templates[st.TemplateID]
		if tpl == nil {
			continue
		}
		templateStrategies[st.TemplateID] = append(templateStrategies[st.TemplateID], st.ID)
	}
	for _, strategyIDs := range templateStrategies {
		sort.Sort(sort.IntSlice(strategyIDs))
	}

	// group:1 - n:template - m:strategies
	groupStrategies := make(map[int][]int, len(da.GroupTemplates))
	for groupID, templates := range da.GroupTemplates {
		groupName := da.GroupNames[groupID]
		if groupName == "" {
			continue
		}
		for _, tplID := range templates {
			groupStrategies[groupID] = append(groupStrategies[groupID], templateStrategies[tplID]...)
		}
	}
	for groupID, strategyIDs := range groupStrategies {
		strategyIDs = utils.IntSliceUnique(strategyIDs)
		sort.Sort(sort.IntSlice(strategyIDs))
		groupStrategies[groupID] = strategyIDs
	}

	hostGroupKeys := make(map[string]string, len(da.HostNames))
	groupKeyStrategies := make(map[string][]int, len(da.GroupHosts))
	groupPlugins := make(map[string][]string, len(da.GroupHosts))
	for hostID, groupIDs := range da.GroupHosts {
		name := da.HostNames[hostID]
		if name == "" {
			// log.Warn("host-no-name", "id", hostID)
			continue
		}
		key := splitToString(groupIDs, "-")
		hostGroupKeys[name] = key
		if _, ok := groupKeyStrategies[key]; !ok {
			var ss []int
			for _, groupID := range groupIDs {
				ss = append(ss, groupStrategies[groupID]...)
			}
			ss = utils.IntSliceUnique(ss)
			sort.Sort(sort.IntSlice(ss))
			groupKeyStrategies[key] = ss
		}
		if _, ok := groupPlugins[key]; !ok {
			var plugins []string
			for _, groupID := range groupIDs {
				plugins = append(plugins, da.GroupPlugins[groupID]...)
			}
			plugins = utils.StringSliceUnique(plugins)
			sort.Sort(sort.StringSlice(plugins))
			groupPlugins[key] = plugins
		}
	}

	eps.HostGroupKeys = hostGroupKeys
	eps.GroupStrategy = groupKeyStrategies
	eps.GroupPlugins = groupPlugins
	eps.Strategies = make(map[int]*models.Strategy, len(da.Strategies))
	eps.HostConfigs = da.HostConfigs
	// simplify strategy data
	for id, s := range da.Strategies {
		eps.Strategies[id] = s.ToSimple()
	}
	eps.HostMaintians = da.HostMaintains
	b, _ := json.Marshal(eps)
	eps.CRC = crc32.ChecksumIEEE(b)
	log.Debug("group-key-strategies", "keys", len(groupKeyStrategies), "crc", eps.CRC)
	da.endpoints = eps
}

func splitToString(a []int, sep string) string {
	if len(a) == 0 {
		return ""
	}
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}
	return strings.Join(b, sep)
}

// EndpointLiveStat is status of endpoint heartbeat living time
type EndpointLiveStat struct {
	Endpoint string `json:"endpoint" db:"hostname"`
	LiveAt   int64  `json:"live_at" db:"live_at"`
	Duration int64  `json:"duration" db:"-"`
}

var (
	noLiveSQL = "SELECT hostname,live_at FROM host WHERE live_at >= %d and live_at <= %d ORDER BY live_at ASC"
)

// NoLiveEndpoints query not living endpoints
// time range is now - begin and now - end
func NoLiveEndpoints(begin int64, end int64) ([]EndpointLiveStat, error) {
	n := time.Now().Unix()
	sqlStr := fmt.Sprintf(noLiveSQL, n-end, n-begin)
	rows, err := portalDB.Queryx(sqlStr)
	if err != nil {
		return nil, err
	}
	var endpoints []EndpointLiveStat
	for rows.Next() {
		ep := EndpointLiveStat{}
		if err = rows.StructScan(&ep); err != nil {
			return nil, err
		}
		if ep.Endpoint != "" {
			ep.Duration = n - ep.LiveAt
			endpoints = append(endpoints, ep)
		}
	}
	return endpoints, nil
}

// GetEndpointsMaintain get maintained endpoints
func GetEndpointsMaintain() map[string]int64 {
	if cachedData == nil {
		return nil
	}
	return cachedData.HostMaintains
}

var epheartbeatSQL = "SELECT live_at FROM host WHERE hostname = ?"

// GetEndpointHeartbeatLive gets live_at of one endpoint
func GetEndpointHeartbeatLive(endpoint string) (int64, error) {
	row := portalDB.QueryRowx(epheartbeatSQL, endpoint)
	var liveAt int64
	err := row.Scan(&liveAt)
	return liveAt, err
}

package sqldata

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

var insertHeartbeatSQL = "INSERT INTO host(hostname, ip, agent_version, plugin_version,live_at,created_at) VALUES ('%s', '%s', '%s', '%s', %d, %d) ON DUPLICATE KEY UPDATE ip='%s', agent_version='%s', plugin_version='%s', live_at=%d"
var updateHeartbeatSQL = "UPDATE host SET ip = ?, agent_version=?, plugin_version=?, live_at=?, remote_ip=? WHERE hostname = ?"
var updateHeartTimeSQL = "UPDATE host SET live_at = %d WHERE hostname IN (%s)"

var (
	heartbeatDurationCount   = expvar.NewAverage("cache.heartbeat_duration", 20)
	heartbeatUpdateFailCount = expvar.NewDiff("cache.heartbeat_fail")
)

func init() {
	expvar.Register(heartbeatDurationCount, heartbeatUpdateFailCount)
}

// UpdateHeartbeat updates heartbeat info to db
func UpdateHeartbeat(heartbeats map[string]models.EndpointHeartbeat) {
	hostInfos := cachedData.HostInfos
	if len(hostInfos) == 0 {
		log.Warn("heartbeat-no-infos")
		return
	}

	nowT := time.Now()
	now := nowT.Unix()
	justUpdateTime := make([]string, 0, len(heartbeats))

	for ep, heart := range heartbeats {
		info, ok := hostInfos[ep]
		if !ok {
			sql := fmt.Sprintf(insertHeartbeatSQL, ep, heart.IP, heart.Version, heart.PluginVersion, now, now,
				heart.IP, heart.Version, heart.PluginVersion, now)
			if _, err := portalDB.Exec(sql); err != nil {
				log.Warn("insert-heart-error", "error", err, "ep", ep, "sql", sql)
				heartbeatUpdateFailCount.Incr(1)
			} else {
				log.Info("insert-heart-new", "ep", ep)
			}
			continue
		}
		info2 := strings.TrimSpace(heart.IP + heart.Version + heart.PluginVersion + heart.Remote)
		if info != info2 {
			if _, err := portalDB.Exec(updateHeartbeatSQL, heart.IP, heart.Version, heart.PluginVersion, now, heart.Remote, ep); err != nil {
				log.Warn("update-heart-value-error", "error", err, "ep", ep)
				heartbeatUpdateFailCount.Incr(1)
			} else {
				log.Debug("update-heart-value", "heart", heart, "ep", ep)
			}
			continue
		}
		justUpdateTime = append(justUpdateTime, strconv.Quote(ep))
	}
	if len(justUpdateTime) > 0 {
		sql := fmt.Sprintf(updateHeartTimeSQL, now, strings.Join(justUpdateTime, ","))
		if _, err := portalDB.Exec(sql); err != nil {
			log.Warn("update-heart-time-error", "error", err, "len", len(justUpdateTime))
			heartbeatUpdateFailCount.Incr(int64(len(justUpdateTime)))
		} else {
			du := utils.DurationSinceMS(nowT)
			heartbeatDurationCount.Set(du)
			log.Debug("update-heartbeat-time", "len", len(justUpdateTime), "du", du)
		}
	}
}

var (
	updateHostServiceUpdatedAtSQL = "UPDATE host_service SET update_at=? WHERE hostname=? AND service_name=?"
	updateHostServiceSQL          = "UPDATE host_service SET updated_at=?,ip=?,remote_ip=?,service_name=?,service_version=?,service_build=? WHERE hostname=? AND service_name=?"
	insertHostServiceSQL          = "INSERT INTO host_service(hostname,ip,remote_ip,service_name,service_version,service_build,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)"

	hostServicesLock sync.RWMutex
)

// UpdateHostService updates host services
func UpdateHostService(service *models.HostService, remoteIP string) {
	if cachedData == nil {
		return
	}
	hostServicesLock.RLock()
	defer hostServicesLock.RUnlock()

	nowUnix := time.Now().Unix()
	oldSvc := cachedData.HostServices[service.Key()]
	if oldSvc == nil {
		if _, err := portalDB.Exec(insertHostServiceSQL, service.Hostname, service.IP, remoteIP, service.ServiceName, service.ServiceVersion, service.ServiceBuild, nowUnix, nowUnix); err != nil {
			log.Warn("insert-hostservice-error", "error", err, "hs", service, "remote", remoteIP)
			return
		}
		log.Info("insert-hostservice", "hs", service, "remote", remoteIP)
		return
	}
	if oldSvc.ValuesString() != service.ValuesString() {
		if _, err := portalDB.Exec(updateHostServiceSQL, nowUnix, service.IP, remoteIP, service.ServiceName, service.ServiceVersion, service.ServiceBuild,
			service.Hostname, service.ServiceName); err != nil {
			log.Warn("update-hostservice-values-error", "error", err, "hs", service, "remote", remoteIP)
			return
		}
		log.Info("update-hostservice", "hs", service, "remote", remoteIP)
		return
	}
	if _, err := portalDB.Exec(updateHostServiceUpdatedAtSQL, nowUnix, service.Hostname, service.ServiceName); err != nil {
		log.Warn("update-hostservice-error", "error", err, "hs", service, "remote", remoteIP)
		return
	}
	log.Info("update-hostservice", "hs", service, "remote", remoteIP)
}

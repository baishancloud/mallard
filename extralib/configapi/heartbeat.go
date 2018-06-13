package configapi

import (
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	heartbeatSyncOnceCount = expvar.NewBase("csdk.heartbeat_once")
)

func init() {
	expvar.Register(heartbeatSyncOnceCount)
	registerFactory("heartbeat", syncHeartbeat)
}

var (
	heartbeatEndpoints = make(map[string]models.EndpointHeartbeat)
	heartbeatLock      sync.RWMutex
)

func syncHeartbeat() {
	heartbeatLock.Lock()
	currents := heartbeatEndpoints
	heartbeatEndpoints = make(map[string]models.EndpointHeartbeat, len(currents))
	heartbeatLock.Unlock()

	if len(currents) == 0 {
		log.Info("heartbeat-0")
		return
	}

	resp, err := httputil.PostJSON(centerAPI+"/api/ping", time.Second*10, currents)
	if err != nil {
		log.Warn("heartbeat-error", "error", err, "heatbeats", len(currents))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Warn("heatbeat-bad-status", "status", resp.StatusCode)
		return
	}
	heartbeatSyncOnceCount.Set(int64(len(currents)))
	log.Info("heartbeat-ok", "len", len(currents))
}

// SetHeartbeat sets heart beat info to sync to config-center
func SetHeartbeat(endpoint, version, plugin, ip, remote string) {
	heartbeatLock.Lock()
	heartbeatEndpoints[endpoint] = models.EndpointHeartbeat{
		Version:       version,
		PluginVersion: plugin,
		IP:            ip,
		Remote:        remote,
		Endpoint:      ""}
	heartbeatLock.Unlock()
}

package configapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	heartbeatSyncOnceCount = expvar.NewBase("csdk.heartbeat_once")
)

const (
	TypeSyncHeartbeat   = "heartbeat"
	TypeSyncHostService = "sync-hostservice"
)

func init() {
	expvar.Register(heartbeatSyncOnceCount)
	registerFactory(TypeSyncHeartbeat, syncHeartbeat)
	registerFactory(TypeSyncHostService, syncHostService)
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

type (
	// HeartbeatResult is result of heartbeat stats request
	HeartbeatResult struct {
		Endpoints []HeartbeatStat `json:"endpoints"`
	}
	// HeartbeatStat is items of heartbeat
	HeartbeatStat struct {
		Endpoint string `json:"endpoint" db:"hostname"`
		LiveAt   int64  `json:"live_at" db:"live_at"`
		Duration int64  `json:"duration" db:"-"`
		key      string
	}
)

// GetHeartbeatData gets current heartbeat data
func GetHeartbeatData(diff int64, timeRange int64) (*HeartbeatResult, error) {
	address := fmt.Sprintf(centerAPI+"/api/endpoint/live?duration=%d&range=%d", diff, timeRange)
	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := new(HeartbeatResult)
	return result, json.Unmarshal(body, result)
}

var (
	currentHostService *models.HostService
)

// SetHostService sets host services to sync
// only support one service in once
func SetHostService(svc *models.HostService) {
	currentHostService = svc
}

func syncHostService() {
	if currentHostService == nil {
		return
	}
	resp, err := httputil.PostJSON(centerAPI+"/api/ping/hostservice", time.Second*10, currentHostService)
	if err != nil {
		log.Warn("sync-hostservice-error", "error", err, "hs", currentHostService)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Warn("sync-hostservice-bad-status", "status", resp.StatusCode)
		return
	}
	log.Info("sync-hostservice-ok", "hs", currentHostService)
}

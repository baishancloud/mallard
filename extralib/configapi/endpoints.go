package configapi

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/baishancloud/mallard/componentlib/center/sqldata"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	cacheEndpoints = new(sqldata.Endpoints)

	hostsInfos   = make(map[string][]interface{})
	hostsConfigs = make(map[string]*models.HostConfig)
	hostsLock    sync.RWMutex
)

const (
	// TypeEndpoints is request type of endpoints data
	TypeEndpoints = "endpoints"
	// TypeHostInfos is request type of request infos
	TypeHostInfos = "hostinfos"
	// TypeHostConfigs is request type of request configs
	TypeHostConfigs = "hostconfigs"
)

func init() {
	registerFactory(TypeEndpoints, reqEndpoints)
	registerFactory(TypeHostInfos, reqHostInfos)
	registerFactory(TypeHostConfigs, reqHostConfigs)
}

func reqEndpoints() {
	url := centerAPI + "/api/endpoints?gzip=1&crc=" + strconv.FormatUint(uint64(cacheEndpoints.CRC), 10)
	eps := new(sqldata.Endpoints)
	statusCode, err := httputil.GetJSON(url, time.Second*5, eps)
	if err != nil {
		log.Warn("req-endpoints-error", "error", err)
		return
	}
	if statusCode == 304 {
		log.Info("req-endpoints-304")
		return
	}
	if eps.CRC != cacheEndpoints.CRC {
		eps.BuildAll()
		cacheEndpoints = eps
		log.Info("req-endpoints-ok", "crc", cacheEndpoints.CRC)
		return
	}
	log.Info("req-endpoints-same", "crc", cacheEndpoints.CRC)
}

func reqHostConfigs() {
	url := centerAPI + "/api/host/configs?gzip=1"
	hosts := make(map[string]*models.HostConfig)
	_, err := httputil.GetJSON(url, time.Second*10, &hosts)
	if err != nil {
		log.Warn("req-hostconfigs-error", "error", err)
		return
	}
	hostsLock.Lock()
	hostsConfigs = hosts
	log.Info("req-hostconfigs", "hosts", len(hosts))
	hostsLock.Unlock()
}

// EndpointConfig gets one endpoint config from cached data
func EndpointConfig(endpoint string) *models.EndpointConfig {
	return cacheEndpoints.Endpoint(endpoint)
}

func reqHostInfos() {
	url := centerAPI + "/api/endpoints/info?gzip=1"
	hosts := make(map[string][]interface{})
	_, err := httputil.GetJSON(url, time.Second*10, &hosts)
	if err != nil {
		log.Warn("req-hostinfos-error", "error", err)
		return
	}
	hostsLock.Lock()
	hostsInfos = hosts
	log.Info("req-hostsinfo", "hosts", len(hosts))
	hostsLock.Unlock()
}

// GetEndpointSertypes gets endpoint sertypes from cache
func GetEndpointSertypes(endpoint string) string {
	hostsLock.RLock()
	defer hostsLock.RUnlock()
	args := hostsInfos[endpoint]
	if len(args) < 6 {
		return ""
	}
	return fmt.Sprint(args[5])
}

// GetLivedHostInfos gets living hosts after lastTime
func GetLivedHostInfos(lastTime int64) map[string][]interface{} {
	hostsLock.RLock()
	defer hostsLock.RUnlock()

	results := make(map[string][]interface{}, len(hostsInfos))
	for name, args := range hostsInfos {
		if len(args) < 5 {
			continue
		}
		t, ok := args[4].(float64)
		if !ok {
			continue
		}
		if int64(t) > lastTime {
			results[name] = hostsInfos[name]
		}
	}
	return results
}

// GetAllHostInfos gets all cached hosts info
func GetAllHostInfos() map[string][]interface{} {
	return hostsInfos
}

// CheckAgentStatus checks agents status
func CheckAgentStatus(endpoint string, status int) bool {
	hostsLock.RLock()
	defer hostsLock.RUnlock()
	cfg := hostsConfigs[endpoint]
	if cfg == nil {
		return false
	}
	return cfg.AgentStatus == status
}

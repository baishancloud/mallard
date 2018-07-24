package configapi

import (
	"strconv"
	"sync"
	"time"

	"github.com/baishancloud/mallard/componentlib/center/sqldata"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	cacheEndpoints = new(sqldata.Endpoints)

	hostsInfos = make(map[string][]interface{})
	hostsLock  sync.RWMutex
)

func init() {
	registerFactory("endpoints", reqEndpoints)
	registerFactory("hostinfos", reqHostInfos)
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

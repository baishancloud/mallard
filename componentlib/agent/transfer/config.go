package transfer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/componentlib/agent/plugins"
	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

var (
	cacheEpData = new(models.EndpointData)

	configReqCount    = expvar.NewDiff("poster.config_req")
	configFailCount   = expvar.NewDiff("poster.config_fail")
	configChangeCount = expvar.NewDiff("poster.config_change")
)

func init() {
	expvar.Register(configFailCount, configReqCount, configChangeCount)
}

// SyncConfig starts config data syncing
func SyncConfig(interval time.Duration, fn func(data *models.EndpointData, isUpdate bool)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	func() {
		for {
			epData, err := getConfig()
			if err != nil {
				log.Warn("req-config-fail", "error", err)
			} else {
				log.Info("req-config-ok", "hash", epData.Hash, "tfr_time", epData.Time)
				isUpdate := false
				if cacheEpData.Hash != epData.Hash {
					cacheEpData = epData
					isUpdate = true
					configChangeCount.Incr(1)
				}
				if fn != nil {
					fn(cacheEpData, isUpdate)
				}
			}
			<-ticker.C
		}
	}()
}

func getConfig() (*models.EndpointData, error) {
	if atomic.LoadInt64(&stopFlag) > 0 {
		log.Warn("config-stopped")
		return cacheEpData, nil
	}
	configReqCount.Incr(1)
	var (
		err      error
		resp     *http.Response
		duration time.Duration
		m        = map[string]string{
			"Agent-Version":  "version",
			"Agent-Plugin":   plugins.Version(),
			"Agent-IP":       serverinfo.IP(),
			"Agent-Endpoint": serverinfo.Hostname(),
		}
	)
	svrInfo := serverinfo.Cached()
	if svrInfo != nil {
		m["Agent-EndpointConf"] = svrInfo.HostnameAllConf
		m["Agent-Endpoint-ID"] = svrInfo.HostIDAllConf
	}
	for i := 0; i < 5; i++ {

		urlLock.RLock()
		idx := urlLatency.Get()
		url := urlList[idx] + urlSuffix["config"]
		urlLock.RUnlock()

		url += "?ep=" + serverinfo.Hostname() + "&hash=" + cacheEpData.Hash + "&gzip=1"
		resp, duration, err = tfrClient.GET(url, m)
		if err != nil {
			// urlLatency.SetFail(idx)
			log.Warn("req-config-once-error", "url", url, "error", err)
			continue
		}
		if resp.StatusCode == 304 {
			log.Debug("req-config-304", "ds", duration.Nanoseconds()/1e6)
			resp.Body.Close()

			transferTimeStr := resp.Header.Get("Transfer-Time")
			transferTime, _ := strconv.ParseInt(transferTimeStr, 10, 64)
			cacheEpData.Time = transferTime

			return cacheEpData, nil
		}
		if resp.StatusCode >= 400 {
			body, _ := ioutil.ReadAll(resp.Body)
			err = fmt.Errorf("bad status %d, %s", resp.StatusCode, body)
			resp.Body.Close()
			log.Debug("req-config-once-error", "url", url, "error", err)
			continue
		}

		transferTimeStr := resp.Header.Get("Transfer-Time")
		transferTime, _ := strconv.ParseInt(transferTimeStr, 10, 64)

		ep := &models.EndpointData{
			Config: &models.EndpointConfig{
				Builtin: &models.EndpointBuiltin{},
			},
			Time: transferTime,
		}
		if err := utils.UngzipJSON(resp.Body, ep); err != nil {
			resp.Body.Close()
			log.Debug("req-config-once-error", "url", url, "error", err)
			continue
		}
		resp.Body.Close()
		log.Debug("req-config-ok", "ds", duration.Nanoseconds()/1e6)
		return ep, nil
	}
	if err != nil {
		configFailCount.Incr(1)
	}
	return nil, err
}

// EndpointData returns current cached models.EndpointData from transfer config requests
func EndpointData() *models.EndpointData {
	return cacheEpData
}

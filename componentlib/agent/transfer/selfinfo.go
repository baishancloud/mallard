package transfer

import (
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/componentlib/agent/plugins"
	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/corelib/utils"
)

// SyncSelfInfo starts sending self info in loops
func SyncSelfInfo(cfgData interface{}) {
	time.Sleep(time.Minute) // do not send directly
	utils.TickerThen(time.Hour*3, func() {
		SendSelfInfo(cfgData)
	})
}

// SendSelfInfo sends self info to transfer
func SendSelfInfo(cfgData interface{}) {
	if atomic.LoadInt64(&stopFlag) > 0 {
		log.Warn("selfinfo-stopped")
		return
	}

	value := map[string]interface{}{
		"endpoint":   cacheEpData,
		"serverinfo": serverinfo.Cached(),
		"config":     cfgData,
		"plugins":    plugins.FilesHash(),
	}

	urlLock.RLock()
	idx := urlLatency.Get()
	url := urlList[idx] + urlSuffix["self"]
	urlLock.RUnlock()

	resp, du, err := tfrClient.POST(url, value, 0)
	if err != nil {
		log.Debug("latency", "history", urlLatency.History())
		log.Warn("selfinfo-send-error", "url", url, "error", err)
		return
	}
	resp.Body.Close()
	log.Info("selfinfo-send-ok", "url", url, "ms", du.Nanoseconds()/1e6)
}

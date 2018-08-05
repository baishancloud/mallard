package transfer

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log      = zaplog.Zap("transfer")
	stopFlag int64
	sendWg   sync.WaitGroup
)

var (
	urlList    []string
	urlSuffix  map[string]string
	urlLatency *utils.Latency
	urlLock    sync.RWMutex
)

// SetURLs sets urls of transfer
func SetURLs(urls []string, suffix map[string]string) {
	if len(urls) == 0 {
		return
	}
	realURLs := make([]string, len(urls))
	for i := range urls {
		realURLs[i] = strings.TrimSuffix(urls[i], "/") // clean ending slash
	}
	urlLock.Lock()
	urlList = realURLs
	urlSuffix = suffix
	urlLatency = utils.NewLatency(len(realURLs), 1e5)
	log.Debug("set-urls", "urls", realURLs, "suffix", suffix)
	urlLock.Unlock()
}

// Stop stops transfer sending, waits all requests finish
func Stop() {
	atomic.StoreInt64(&stopFlag, 1)
	sendWg.Wait()
	log.Info("stop")
}

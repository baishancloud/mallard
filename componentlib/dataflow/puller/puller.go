package puller

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	urlsUnit    = make(map[string]*Unit)
	urlsLock    sync.RWMutex
	stopFlag    int64
	metricsChan chan<- []*models.Metric

	log = zaplog.Zap("puller")
)

// SetChan sets chan to handle pulling metrics
func SetChan(mCh chan<- []*models.Metric) {
	metricsChan = mCh
}

// SetURLs sets urls to pull
func SetURLs(urlList map[string]string, concurrency int) {
	realURLs := make(map[string]string, len(urlList))
	for key, url := range urlList {
		realURLs[key] = strings.TrimSuffix(url, "/") + "/api/metric_pop"
	}
	now := time.Now().Unix()
	urlsLock.Lock()
	for key, url := range realURLs {
		unit := urlsUnit[key]
		if unit == nil {
			unit = NewUnit(key, url, concurrency)
			urlsUnit[key] = unit
			log.Info("add-url-unit", "key", key)
		}
		unit.SetURL(url)
		unit.createTime = now
		unit.Start(concurrency)
	}
	for key := range urlsUnit {
		unit := urlsUnit[key]
		if unit.createTime < now {
			delete(urlsUnit, key)
			go unit.Stop()
			log.Info("del-url-unit", "key", key)
		}
	}
	log.Info("set-urls", "urls", realURLs)
	urlsLock.Unlock()
}

// Stop stops puller grabing actions
func Stop() {
	atomic.StoreInt64(&stopFlag, 1)
	urlsLock.Lock()
	for _, unit := range urlsUnit {
		unit.Stop()
	}
	urlsLock.Unlock()
	log.Info("stop")
}

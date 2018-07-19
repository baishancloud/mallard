package puller

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/corelib/container"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	urlsUnit     = make(map[string]*Unit)
	urlsLock     sync.RWMutex
	stopFlag     int64
	metricsQueue container.LimitedQueue

	log = zaplog.Zap("puller")
)

var (
	// queueLengthCount   = expvar.NewBase("queue_length")
	queuePushCount     = expvar.NewDiff("queue_push")
	queuePushFailCount = expvar.NewDiff("queue_push_fail")
	/*"puller_queue.cnt": 10426,
	"puller_pop.diff": 526297,
	"puller_batch.cnt": 1749*/
)

func init() {
	expvar.Register(queuePushCount, queuePushFailCount)
}

// SetQueue sets queue to handle pulling metrics
func SetQueue(queue container.LimitedQueue) {
	metricsQueue = queue
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

func SyncExpvars(duration time.Duration, file string) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		<-ticker.C
		genExpvars(file)
	}
}

func genExpvars(file string) {
	results := make(map[string]map[string]interface{})
	urlsLock.RLock()
	for key, unit := range urlsUnit {
		results[key] = unit.counters()
	}
	urlsLock.RUnlock()
	now := time.Now().Unix()
	var metrics []*models.Metric
	for name, result := range results {
		metric := &models.Metric{
			Name:   "mallard2_store_puller",
			Time:   now,
			Fields: result,
			Tags: map[string]string{
				"transfer": name,
			},
		}
		metrics = append(metrics, metric)
	}
	log.Debug("puller-stats", "metrics", metrics)
	if file != "" {
		b, _ := json.Marshal(metrics)
		ioutil.WriteFile(file, b, 0644)
	}
}

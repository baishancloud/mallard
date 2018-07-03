package syscollector

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

// Collector is the function to collect metrics
type Collector func() ([]*models.Metric, error)

var (
	collectorFactory = make(map[string]Collector, 15)
	collectorLock    sync.RWMutex
	collectCounter   = expvar.NewDiff("sys.collect")
)

func init() {
	expvar.Register(collectCounter)
}

func registerFactory(name string, c Collector) {
	collectorLock.Lock()
	collectorFactory[name] = c
	collectorLock.Unlock()
}

var (
	collectorStopFlag uint64
	log               = zaplog.Zap("syscollector")
)

// Collect runs sysycollectors , pushes metric values to channel and pushes error values to channel
func Collect(prefix string, interval time.Duration, metricsChan chan<- []*models.Metric, errorChan chan<- error) {
	log.Info("init", "prefix", prefix, "interval", int(interval.Seconds()))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if atomic.LoadUint64(&collectorStopFlag) > 0 {
			return
		}
		metrics := make([]*models.Metric, 0, 50)
		collectorLock.RLock()
		for key, factory := range collectorFactory {
			if factory == nil {
				continue
			}
			values, err := factory()
			if err != nil {
				errorChan <- fmt.Errorf("%s, %s", key, err.Error())
			}
			if len(values) > 0 {
				metrics = append(metrics, values...)
			}
		}
		collectorLock.RUnlock()
		if len(metrics) > 0 {
			if prefix != "" {
				for _, m := range metrics {
					m.Name = prefix + "." + m.Name
				}
			}
			metricsChan <- metrics
			log.Info("collect", "metrics", len(metrics))
			collectCounter.Incr(int64(len(metrics)))
		}
		<-ticker.C
	}
}

// StopCollect stops collecting metrics
func StopCollect() {
	atomic.StoreUint64(&collectorStopFlag, 1)
}

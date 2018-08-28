package syscollector

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

type (
	// CollectorFunc is function to collect metrics
	CollectorFunc = func() ([]*models.Metric, error)
	// Collector contains the function to collect metrics and step
	Collector struct {
		Func CollectorFunc
		Step int
	}
)

var (
	collectorFactory = make(map[string]Collector, 15)
	collectorLock    sync.RWMutex
	collectCounter   = expvar.NewDiff("sys.collect")
)

func init() {
	expvar.Register(collectCounter)
}

func registerFactory(name string, c interface{}) {
	collectorLock.Lock()
	defer collectorLock.Unlock()
	if ct, ok := c.(Collector); ok {
		collectorFactory[name] = ct
		return
	}
	if cfunc, ok := c.(CollectorFunc); ok {
		collectorFactory[name] = Collector{cfunc, 1}
		return
	}
}

var (
	collectorStopFlag uint64
	log               = zaplog.Zap("syscollector")
)

// Collect runs sysycollectors , pushes metric values to channel and pushes error values to channel
func Collect(prefix string, interval time.Duration, metricsChan chan<- []*models.Metric) {
	log.Info("init", "prefix", prefix, "interval", int(interval.Seconds()))

	// print running system collector
	keys := make([]string, 0, len(collectorFactory))
	collectorLock.RLock()
	for key := range collectorFactory {
		keys = append(keys, key)
	}
	collectorLock.RUnlock()
	sort.Sort(sort.StringSlice(keys))
	log.Info("factory", "keys", keys)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var step int
	for {
		if atomic.LoadUint64(&collectorStopFlag) > 0 {
			return
		}
		metrics := make([]*models.Metric, 0, 100)
		collectorLock.RLock()
		for key, factory := range collectorFactory {
			if factory.Func == nil {
				continue
			}
			if factory.Step > 1 && step%factory.Step != 0 {
				continue
			}
			values, err := factory.Func()
			if err != nil {
				log.Warn("collect-error", "error", fmt.Errorf("%s, %s", key, err.Error()))
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

		// incr step
		step++
		if step == 100 {
			step = 0
		}
		<-ticker.C
	}
}

// StopCollect stops collecting metrics
func StopCollect() {
	atomic.StoreUint64(&collectorStopFlag, 1)
}

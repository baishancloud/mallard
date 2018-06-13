package transfer

import (
	"sync/atomic"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	// MaxMetricsInOnce is max length of metrics in one requests
	MaxMetricsInOnce = 1000

	metricSendCount    = expvar.NewDiff("transfer.metrics")
	metricFailCount    = expvar.NewDiff("transfer.metrics_fail")
	metricLatencyCount = expvar.NewAverage("transfer.metrics_latency", 10)
)

func init() {
	expvar.Register(metricFailCount, metricLatencyCount, metricSendCount)
}

// Metrics sends metrics to transfer
func Metrics(metrics []*models.Metric) {
	if atomic.LoadInt64(&stopFlag) > 0 {
		log.Warn("metrics-stopped")
		return
	}

	dataLen := len(metrics)
	if dataLen == 0 {
		log.Warn("metrics-empty")
		return
	}
	if len(metrics) > MaxMetricsInOnce {
		idx := len(metrics) / 2
		log.Debug("metrics-split", "all", len(metrics), "idx", idx)
		go Metrics(metrics[:idx])
		go Metrics(metrics[idx:])
		return
	}

	sendWg.Add(1)
	defer sendWg.Done()

	var isSend bool
	for i := 0; i <= 3; i++ {

		urlLock.RLock()
		idx := urlLatency.Get()
		url := urlList[idx] + urlSuffix["metric"]
		urlLock.RUnlock()

		resp, du, err := tfrClient.POST(url, metrics, dataLen)
		if err != nil {
			log.Debug("latency", "history", urlLatency.History())
			log.Warn("metrics-send-once-error", "url", url, "error", err)
			urlLatency.SetFail(idx)
			continue
		}
		resp.Body.Close()
		ds := du.Nanoseconds() / 1e6
		ds2 := ds*100/int64(len(metrics)) + 1
		urlLatency.Set(idx, ds2)
		log.Info("metrics-send-ok", "url", url, "len", dataLen, "ms", ds, "ds", ds2)
		metricLatencyCount.Set(ds2)
		isSend = true
		break
	}
	if !isSend {
		log.Warn("metrics-send-fail", "len", dataLen)
		metricFailCount.Incr(int64(len(metrics)))
		return
	}
	metricSendCount.Incr(int64(len(metrics)))
}

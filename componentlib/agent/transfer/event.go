package transfer

import (
	"sync/atomic"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	// MaxEventsInOnce is max length of events in one request
	MaxEventsInOnce = 2000

	eventSendCount    = expvar.NewDiff("poster.event")
	eventFailCount    = expvar.NewDiff("poster.event_fail")
	eventLatencyCount = expvar.NewAverage("poster.event_latency", 10)
)

func init() {
	expvar.Register(eventFailCount, eventLatencyCount, eventSendCount)
}

// Events sends events to transfer
func Events(events []*models.Event) {
	if atomic.LoadInt64(&stopFlag) > 0 {
		log.Warn("events-stopped")
		return
	}

	dataLen := len(events)
	if dataLen == 0 {
		log.Warn("events-empty")
		return
	}
	if len(events) > MaxEventsInOnce {
		idx := len(events) / 2
		log.Debug("events-split", "all", len(events), "idx", idx)
		go Events(events[:idx])
		go Events(events[idx:])
		return
	}

	sendWg.Add(1)
	defer sendWg.Done()

	var isSend bool
	for i := 0; i <= 3; i++ {

		urlLock.RLock()
		idx := urlLatency.Get()
		url := urlList[idx] + urlSuffix["event"]
		urlLock.RUnlock()

		resp, du, err := tfrClient.POST(url, events, dataLen)
		if err != nil {
			log.Debug("latency", "history", urlLatency.History())
			log.Warn("events-send-once-error", "url", url, "error", err)
			continue
		}
		resp.Body.Close()
		ds := du.Nanoseconds() / 1e6
		log.Info("events-send-ok", "url", url, "len", dataLen, "ms", ds)
		eventLatencyCount.Set(ds)
		isSend = true
		break
	}
	if !isSend {
		log.Warn("events-send-fail", "len", dataLen)
		eventFailCount.Incr(int64(len(events)))
		return
	}
	eventSendCount.Incr(int64(len(events)))
}

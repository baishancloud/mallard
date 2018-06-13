package influxdb

import (
	"sync"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	client "github.com/influxdata/influxdb/client/v2"
)

var (
	log = zaplog.Zap("influxdb")
	wg  sync.WaitGroup
)

// Process processes metrics channel data to sending
func Process(mChan <-chan []*models.Metric) {
	for {
		metrics := <-mChan
		if len(metrics) == 0 {
			continue
		}
		wg.Add(1)
		go func(ms []*models.Metric) {
			SendMetrics(metrics)
			wg.Done()
		}(metrics)
	}
}

// SendMetrics sends metrics to influxdb
func SendMetrics(metrics []*models.Metric) {
	dataList := make(map[string][]*client.Point)
	for _, m := range metrics {
		p, err := metric2Point(m)
		if err != nil {
			log.Warn("point-error", "error", err, "metric", m)
			continue
		}
		groupLock.RLock()
		accepts := groupAccept[m.Name]
		if len(accepts) == 0 {
			accepts = groupAccept["*"]
		}
		groupLock.RUnlock()
		if len(accepts) == 0 {
			continue
		}
		for _, name := range accepts {
			dataList[name] = append(dataList[name], p)
		}
	}

	groupLock.RLock()
	defer groupLock.RUnlock()
	for name, points := range dataList {
		if g := groups[name]; g != nil {
			g.Send(points)
		}
	}
}

// Stop stops sending to influxdb
func Stop() {
	wg.Wait()
}

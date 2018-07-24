package influxdb

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/container"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	client "github.com/influxdata/influxdb/client/v2"
)

var (
	log = zaplog.Zap("influxdb")
	wg  sync.WaitGroup

	acceptDumpFile = "metrics_accepts.dump"

	queueLengthCount = expvar.NewBase("queue_length")
	queueBatchCount  = expvar.NewBase("queue_batch")
	queuePopCount    = expvar.NewDiff("queue_pop")
)

func init() {
	expvar.Register(queueBatchCount, queueLengthCount, queuePopCount)
}

// Process processes metrics queue data to sending
func Process(queue container.LimitedQueue) {
	go loopQueue(queue, 2e3, time.Millisecond*100)
}

func changeBatch(batchBase int, qLen int) int {
	ratio := float64(qLen)/float64(batchBase) - 1
	ratio = 1 + ratio/8
	if ratio < 0.85 {
		ratio = 0.85
	}
	if ratio > 3 {
		ratio = 3
	}
	return int(float64(batchBase)*ratio) + 1
}

func changeBatchBase(mCount, count int, batchBase int) int {
	if batchBase > 0 {
		base := mCount / count
		ratio := 1 + (float64(base)/float64(batchBase)-1)/5
		if ratio < 0.85 {
			ratio = 0.85
		}
		if ratio > 3 {
			ratio = 3
		}
		return int(float64(batchBase)*ratio) + 1
	}
	return mCount / count
}

func loopQueue(queue container.LimitedQueue, batch int, interval time.Duration) {
	var (
		count     int
		mCount    int
		batchBase int
		ticker    = time.NewTicker(interval)
	)
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			if batchBase > 0 && count%5 == 1 {
				batch = changeBatch(batchBase, queue.Len())
			}
			if count >= 10 {
				batchBase = changeBatchBase(mCount, count, batchBase)
				mCount = 0
				count = 0
			}
			data := queue.PopBatch(batch)
			if len(data) == 0 {
				continue
			}
			log.Debug("batch", "batch", batch, "base", batchBase, "queue", queue.Len())
			count++
			mCount += len(data)

			metrics := make([]*models.Metric, 0, len(data))
			for _, item := range data {
				if m, ok := item.(*models.Metric); ok {
					metrics = append(metrics, m)
				}
			}
			queuePopCount.Incr(int64(len(metrics)))

			wg.Add(1)
			go func(ms []*models.Metric) {
				SendMetrics(metrics)
				wg.Done()
			}(metrics)

			queueBatchCount.Set(int64(batch))
			queueLengthCount.Set(int64(queue.Len()))
		}
	}()
}

// SendMetrics sends metrics to influxdb
func SendMetrics(metrics []*models.Metric) {
	addonsAccepting := make(map[string][]string)
	dataList := make(map[string][]*client.Point)
	nowUnix := time.Now().Unix()
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
			addonsAccepting[m.Name] = accepts
		}
		groupLock.RUnlock()
		if len(accepts) == 0 {
			continue
		}
		for _, name := range accepts {
			expire := groupExpire[name]
			if expire > 0 && nowUnix-m.Time >= expire {
				log.Warn("point-expire", "metric", m, "group", name)
				continue
			}
			dataList[name] = append(dataList[name], p)
		}
	}

	groupLock.RLock()
	for name, points := range dataList {
		if g := groups[name]; g != nil {
			g.Send(points)
		}
	}
	groupLock.RUnlock()

	if len(addonsAccepting) > 0 {
		groupLock.Lock()
		for name, accepts := range addonsAccepting {
			groupAccept[name] = accepts
		}
		b, _ := json.Marshal(groupAccept)
		ioutil.WriteFile(acceptDumpFile, b, 0644)
		log.Info("add-accepts", "count", len(groupAccept))
		groupLock.Unlock()
	}
}

func SyncExpvars(interval time.Duration, file string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		genExpvars(file)
	}
}

func genExpvars(file string) {
	now := time.Now().Unix()
	groupLock.RLock()
	var metrics []*models.Metric
	for name, group := range groups {
		nodeResults := group.counters()
		for nodeName, result := range nodeResults {
			metric := &models.Metric{
				Name:   "mallard2_store_influxdb",
				Time:   now,
				Fields: result,
				Tags: map[string]string{
					"node":    nodeName,
					"cluster": name,
				},
			}
			metrics = append(metrics, metric)
		}
	}
	groupLock.RUnlock()

	log.Debug("influxdb-stats", "metrics", metrics)
	if file != "" {
		b, _ := json.Marshal(metrics)
		ioutil.WriteFile(file, b, 0644)
	}
}

// Stop stops sending to influxdb
func Stop() {
	wg.Wait()
}

package puller

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/componentlib/dataflow/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

const (
	// ErrorWaitDuration is waiting time after serveral errors
	ErrorWaitDuration = time.Second * 30
	// NextWaitDuration is waiting time after 204 response or error
	NextWaitDuration = time.Second * 5
)

// Unit is pulling unit for one url
type Unit struct {
	key           string
	url           string
	createTime    int64
	isErrorStatus int

	startFlag int64
	wg        sync.WaitGroup

	failCount      *expvar.DiffMeter
	reqCount       *expvar.DiffMeter
	zeroCounter    *expvar.DiffMeter
	okCount        *expvar.DiffMeter
	metricCounter  *expvar.DiffMeter
	latencyCounter *expvar.AvgMeter
}

// NewUnit creates new unit with keyname and url
// it starts pulling auto
func NewUnit(key, url string, concurrency int) *Unit {
	unit := &Unit{
		key:            key,
		url:            url,
		failCount:      expvar.NewDiff("fail"),
		reqCount:       expvar.NewDiff("req"),
		zeroCounter:    expvar.NewDiff("zero"),
		okCount:        expvar.NewDiff("ok"),
		metricCounter:  expvar.NewDiff("metric"),
		latencyCounter: expvar.NewAverage("latency", 50),
	}
	unit.Start(concurrency)
	return unit
}

func (u *Unit) loop() {
	defer u.wg.Done()
	for {
		if atomic.LoadInt64(&u.startFlag) < 0 {
			log.Info("loop-stop", "key", u.key)
			return
		}
		if u.isErrorStatus >= 50 {
			time.Sleep(time.Second * 30)
		}
		u.wg.Add(1)
		u.reqCount.Incr(1)
		resp, duration, err := getURL(u.url, time.Second*20)
		if err != nil {
			log.Warn("req-url-error", "key", u.key, "error", err)
			u.isErrorStatus++
			u.wg.Done()
			time.Sleep(NextWaitDuration)
			u.failCount.Incr(1)
			continue
		}
		if resp.StatusCode >= 400 {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Warn("req-url-bad-status", "key", u.key, "status", resp.StatusCode, "body", string(body))
			resp.Body.Close()
			u.isErrorStatus++
			u.wg.Done()
			continue
		}
		if resp.StatusCode == 204 {
			u.isErrorStatus = 0
			log.Debug("req-url-204", "key", u.key)
			u.wg.Done()
			time.Sleep(NextWaitDuration)
			u.zeroCounter.Incr(1)
			continue
		}
		u.isErrorStatus = 0
		if err := u.parsePullResponse(resp); err != nil {
			u.failCount.Incr(1)
		} else {
			u.okCount.Incr(1)
		}
		resp.Body.Close()
		du := utils.DurationMS(duration)
		log.Debug("req-url-ok", "key", u.key, "ms", du)
		u.latencyCounter.Set(du)
		u.wg.Done()
	}
}

func (u *Unit) parsePullResponse(resp *http.Response) error {
	if resp.Header.Get("Data-Type") == "pack" {
		return u.parseResponsePacket(resp)
	}
	return u.parseResponseMetrics(resp)
}

func (u *Unit) parseResponseMetrics(resp *http.Response) error {
	dataLen, _ := strconv.Atoi(resp.Header.Get("Data-Length"))
	metrics := make([]*models.Metric, 0, dataLen)
	if err := utils.UngzipJSON(resp.Body, &metrics); err != nil {
		log.Warn("req-metrics-error", "key", u.key, "error", err, "data-len", dataLen)
		return err
	}
	log.Debug("pull-ok", "key", u.key, "len", len(metrics), "data-len", dataLen)
	if metricsQueue != nil && len(metrics) > 0 {
		u.metricCounter.Incr(int64(len(metrics)))
		queuePushCount.Incr(int64(len(metrics)))
		values := make([]interface{}, len(metrics))
		for i := range metrics {
			values[i] = metrics[i]
		}
		if !metricsQueue.PushBatch(values) {
			queuePushFailCount.Incr(int64(len(metrics)))
		}
	}
	return nil
}

func (u *Unit) parseResponsePacket(resp *http.Response) error {
	dataLen, _ := strconv.Atoi(resp.Header.Get("Data-Length"))
	packs, err := queues.PacketsFromReader(resp.Body, dataLen)
	if len(packs) > 0 {
		metrics, err := packs.ToMetrics()
		if err != nil {
			log.Warn("req-pack-metrics-error", "key", u.key, "error", err, "data-len", dataLen, "packs", len(packs))
			return err
		}
		log.Debug("pull-ok", "key", u.key, "len", len(metrics), "data-len", dataLen)
		if metricsQueue != nil && len(metrics) > 0 {
			u.metricCounter.Incr(int64(len(metrics)))
			queuePushCount.Incr(int64(len(metrics)))
			values := make([]interface{}, len(metrics))
			for i := range metrics {
				values[i] = metrics[i]
			}
			if !metricsQueue.PushBatch(values) {
				queuePushFailCount.Incr(int64(len(metrics)))
			}
		}
	}
	if err != nil {
		log.Warn("req-parse-error", "key", u.key, "error", err, "data-len", dataLen, "packs", len(packs))
		return err
	}
	return nil
}

// SetURL sets url to the unit
func (u *Unit) SetURL(url string) {
	if u.url == url {
		return
	}
	u.url = url
}

// Start starts loop
func (u *Unit) Start(concurrency int) {
	if atomic.LoadInt64(&u.startFlag) > 0 {
		log.Debug("unit-had-started", "key", u.key)
		return
	}
	atomic.StoreInt64(&u.startFlag, 1)
	log.Info("unit-start", "key", u.key, "concurrency", concurrency)
	for i := 0; i < concurrency; i++ {
		u.wg.Add(1)
		go u.loop()
	}
}

// Stop stops unit pulling action
func (u *Unit) Stop() {
	atomic.StoreInt64(&u.startFlag, -1)
	if u.isErrorStatus == 0 {
		u.wg.Wait()
	}
	log.Info("unit-stop", "key", u.key, "error_status", u.isErrorStatus)
}

func (u *Unit) counters() map[string]interface{} {
	values := []interface{}{u.failCount, u.reqCount, u.zeroCounter, u.okCount, u.metricCounter, u.latencyCounter}
	return expvar.ExposeFactory(values, false)
}

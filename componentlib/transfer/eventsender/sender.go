package eventsender

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
	MaxIdleConns:        300,
	MaxIdleConnsPerHost: 100,
}

var (
	eventSendCount        = expvar.NewQPS("eventor.send")
	eventSendEventsCount  = expvar.NewQPS("eventor.send_events")
	eventSendFailCount    = expvar.NewDiff("eventor.send_fail")
	eventQueueLengthCount = expvar.NewBase("eventor.queue_length")
	eventSendOnceAvg      = expvar.NewAverage("event.send_avg", 50)
)

func init() {
	expvar.Register(eventQueueLengthCount, eventSendEventsCount, eventSendCount, eventSendFailCount, eventSendOnceAvg)
}

var (
	urls     = make(map[string]string)
	urlsLock sync.RWMutex
	stopFlag int64
	wg       sync.WaitGroup

	log = zaplog.Zap("eventSender")
)

// SetURLs sets urls
func SetURLs(rawURLs map[string]string) {
	readURLs := make(map[string]string, len(rawURLs))
	for key, u := range rawURLs {
		readURLs[key] = strings.TrimSuffix(u, "/") + "/api/event"
	}
	urlsLock.Lock()
	urls = readURLs
	urlsLock.Unlock()
}

// ProcessQueue handle queues to sender
func ProcessQueue(queue *queues.Queue, batch int) {
	for {
		if atomic.LoadInt64(&stopFlag) > 0 {
			log.Info("stop")
			return
		}
		if !popAndSend(queue, batch) {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func popAndSend(q *queues.Queue, batch int) bool {
	packets, err := q.Pop(batch)
	if err != nil {
		log.Warn("pop-error", "error", err)
		return false
	}
	if len(packets) == 0 {
		return false
	}
	wg.Add(1)
	sendValues(packets)
	return true
}

func sendValues(packets queues.Packets) {
	defer wg.Done()

	dataLen := packets.DataLen()
	data, err := json.Marshal(packets)
	if err != nil {
		log.Warn("json-error", "error", err)
		return
	}
	eventSendCount.Incr(1)
	eventSendEventsCount.Incr(int64(dataLen))
	eventSendOnceAvg.Set(int64(dataLen))

	urlsLock.RLock()
	defer urlsLock.RUnlock()

	for key, url := range urls {
		sendOnce(key, url, data, dataLen, 2)
	}
}

func sendOnce(urlKey, url string, data []byte, dataLen int, retry int) {
	request, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Warn("new-req-error", "url", url, "error", err)
		return
	}
	request.Header.Add("Content-Type", "application/m-pack")
	request.Header.Add("Data-Length", strconv.Itoa(dataLen))
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Warn("send-once-error", "url", url, "error", err)
		if retry > 0 {
			sendOnce(urlKey, url, data, dataLen, retry-1)
		} else {
			eventSendFailCount.Incr(int64(dataLen))
		}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Warn("send-bad-status", "status", resp.StatusCode, "url", url, "error", err)
		if retry > 0 {
			sendOnce(urlKey, url, data, dataLen, retry-1)
		} else {
			eventSendFailCount.Incr(int64(dataLen))
		}
		return
	}
	log.Debug("send-ok", "bytes", len(data), "len", dataLen, "to", urlKey)
}

// Stop stops event sender
func Stop() {
	atomic.StoreInt64(&stopFlag, 1)
	wg.Wait()
}

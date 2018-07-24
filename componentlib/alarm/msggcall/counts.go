package msggcall

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	userCounts     = make(map[string]*msggUserCount)
	userCountsLock sync.RWMutex
)

type msggUserCount struct {
	User   string
	Method string
	Count  *expvar.DiffMeter
}

func calculateUserCount(req *msggRequest) {
	userCountsLock.Lock()
	defer userCountsLock.Unlock()
	if req.SendRequest == nil {
		return
	}
	for _, user := range req.SendRequest.Emails {
		key := "email-" + user
		cnt := userCounts[key]
		if cnt == nil {
			userCounts[key] = &msggUserCount{
				User:   user,
				Method: "email",
				Count:  expvar.NewDiff("msgg"),
			}
		}
		userCounts[key].Count.Incr(1)
	}

	for _, user := range req.SendRequest.Wechats {
		key := "wechat-" + user
		cnt := userCounts[key]
		if cnt == nil {
			userCounts[key] = &msggUserCount{
				User:   user,
				Method: "wechat",
				Count:  expvar.NewDiff("msgg"),
			}
		}
		userCounts[key].Count.Incr(1)
	}

	for _, user := range req.SendRequest.SMSs {
		key := "sms-" + user
		cnt := userCounts[key]
		if cnt == nil {
			userCounts[key] = &msggUserCount{
				User:   user,
				Method: "sms",
				Count:  expvar.NewDiff("msgg"),
			}
		}
		userCounts[key].Count.Incr(1)
	}

	for _, user := range req.SendRequest.Phones {
		key := "phone-" + user
		cnt := userCounts[key]
		if cnt == nil {
			userCounts[key] = &msggUserCount{
				User:   user,
				Method: "phone",
				Count:  expvar.NewDiff("msgg"),
			}
		}
		userCounts[key].Count.Incr(1)
	}
}

var (
	countMetric string
)

// SyncPrintCount starts prints counters to file
func SyncPrintCount(metricName string, interval time.Duration, file string) {
	if metricName == "" || interval < time.Second {
		return
	}
	log.Debug("init-user-counts", "interval", int(interval.Seconds()), "file", file)
	countMetric = metricName
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		metrics := userCountsMetrics()
		log.Info("user-counts", "counts", len(metrics))
		if file != "" {
			b, _ := json.Marshal(metrics)
			ioutil.WriteFile(file, b, 0644)
		}
	}
}

func userCountsMetrics() []*models.Metric {
	userCountsLock.RLock()
	defer userCountsLock.RUnlock()

	nowUnix := time.Now().Unix()
	metrics := make([]*models.Metric, 0, len(userCounts))
	for _, count := range userCounts {
		metric := &models.Metric{
			Name:  countMetric,
			Time:  nowUnix,
			Value: float64(count.Count.Diff()),
			Tags: map[string]string{
				"user":   count.User,
				"method": count.Method,
			},
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

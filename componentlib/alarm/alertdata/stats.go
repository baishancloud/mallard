package alertdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	eventStatSQL = "select id,eid,status,strategy_id,duration from %s where event_time >= %d order by id asc"
)

// StatEvent is event data used to stat
type StatEvent struct {
	ID         int    `db:"id"`
	EID        string `db:"eid"`
	Status     string `db:"status"`
	StrategyID int    `db:"strategy_id"`
	Duration   int    `db:"duration"`
}

// GetStatEvents gets all stat events from begin time in proper table
func GetStatEvents(beginTime int64) ([]*StatEvent, error) {
	if clerkDB == nil {
		return nil, ErrDbNil
	}
	statSQL := fmt.Sprintf(eventStatSQL, EventTableName(beginTime), beginTime)
	rows, err := clerkDB.Queryx(statSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*StatEvent
	for rows.Next() {
		event := new(StatEvent)
		if err = rows.StructScan(event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

// StatsOption is option for alarms stats
type StatsOption struct {
	Interval   time.Duration
	TimeRange  int64
	MetricName string
	MetricFile string
	CacheFile  string
}

// ScanStats scans alarm event stats
func ScanStats(opt StatsOption) {
	// check options
	if opt.MetricName == "" || opt.Interval == 0 || opt.TimeRange == 0 {
		log.Warn("stats-invalid-option", "opt", opt)
		return
	}
	log.Info("begin-stat", "diff", opt.TimeRange, "interval", int(opt.Interval.Seconds()), "metric", opt.MetricName, "file", opt.MetricFile)

	time.Sleep(time.Second * 10)
	ticker := time.NewTicker(opt.Interval)
	defer ticker.Stop()
	for {
		st := <-ticker.C

		tUnix := time.Now().Unix() - opt.TimeRange
		events, err := GetStatEvents(tUnix)
		if err != nil {
			log.Warn("get-stats-error", "error", err)
			alertDbErrorCount.Incr(1)
			continue
		}
		olds := readAlarmStatDump(opt.CacheFile)
		log.Debug("read-olds", "len", len(olds))

		results, metrics := checkAlarmStat(events, olds, st.Unix())
		for _, m := range metrics {
			m.Name = opt.MetricName
		}
		log.Debug("check-news", "len", len(results))

		b, _ := json.Marshal(metrics)
		ioutil.WriteFile(opt.MetricFile, b, 0644)

		writeAlarmStatDump(results, opt.CacheFile)
		log.Info("scan-stat-ok", "metrics", len(metrics))
	}
}

// AlarmStatResult is result of alarm event stats
type AlarmStatResult struct {
	StrategyID   int
	OKCount      int
	ProblemCount int
	AllCount     int
	HappenCount  int
}

func checkAlarmStat(
	events []*StatEvent,
	olds map[int]*AlarmStatResult,
	now int64) (map[int]*AlarmStatResult, []*models.Metric) {
	result := make(map[int]*AlarmStatResult)
	for _, event := range events {
		res := result[event.StrategyID]
		if res == nil {
			res = &AlarmStatResult{
				StrategyID: event.StrategyID,
			}
			result[event.StrategyID] = res
		}
		if event.Status == models.EventProblem.String() {
			res.ProblemCount++
		} else {
			res.OKCount++
		}
	}
	var metrics []*models.Metric
	for _, res := range result {
		res.AllCount = res.OKCount + res.ProblemCount
		res.HappenCount = res.ProblemCount - res.OKCount
		m := &models.Metric{
			// Name:  sc.metricName,
			Time:  now,
			Value: float64(res.HappenCount),
			Tags: map[string]string{
				"strategyid": strconv.Itoa(res.StrategyID),
			},
			Fields: map[string]interface{}{
				"all":     res.AllCount,
				"ok":      res.OKCount,
				"problem": res.ProblemCount,
				"happen":  res.HappenCount,
			},
			Step: 60,
		}
		metrics = append(metrics, m)
	}
	otherMetrics := compareOldAlarmStat(result, olds, now)
	if len(otherMetrics) > 0 {
		metrics = append(metrics, otherMetrics...)
	}
	return result, metrics
}

func compareOldAlarmStat(news, olds map[int]*AlarmStatResult, now int64) []*models.Metric {
	var metrics []*models.Metric
	for sid := range olds {
		nv := news[sid]
		if nv == nil {
			m := &models.Metric{
				// Name:  statMetricName,
				Time:  now,
				Value: 0,
				Tags: map[string]string{
					"strategyid": strconv.Itoa(sid),
				},
				Fields: map[string]interface{}{
					"all":     0,
					"ok":      0,
					"problem": 0,
					"happen":  0,
				},
				Step: 30,
			}
			metrics = append(metrics, m)
		}
	}
	return metrics
}

func writeAlarmStatDump(dumps map[int]*AlarmStatResult, file string) {
	b, _ := json.Marshal(dumps)
	ioutil.WriteFile(file, b, 0644)
}

func readAlarmStatDump(file string) map[int]*AlarmStatResult {
	b, _ := ioutil.ReadFile(file)
	if len(b) == 0 {
		return nil
	}
	results := make(map[int]*AlarmStatResult)
	json.Unmarshal(b, &results)
	return results
}

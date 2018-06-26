package multijudge

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
)

type (
	eventItem struct {
		Metric          *models.Metric   `json:"metric,omitempty"`
		MultiStrategyID int              `json:"multi_strategy_id,omitempty"`
		GroupHash       string           `json:"group_hash,omitempty"`
		MetricHash      string           `json:"metric_hash,omitempty"`
		MetricValueHash string           `json:"-"`
		LeftValue       float64          `json:"left_value,omitempty"`
		Score           float64          `json:"score,omitempty"`
		Strategy        *models.Strategy `json:"strategy,omitempty"`
	}
	eventItems struct {
		Items map[string]*eventItem `json:"items,omitempty"`
		lock  sync.RWMutex
	}
)

func (ei *eventItems) Add(item *eventItem) {
	ei.lock.Lock()
	ei.Items[item.MetricValueHash] = item
	ei.lock.Unlock()
}

func (ei *eventItems) Remove(metricValueHash string) bool {
	ei.lock.Lock()
	_, ok := ei.Items[metricValueHash]
	if ok {
		delete(ei.Items, metricValueHash)
	}
	ei.lock.Unlock()
	return ok
}

func (ei *eventItems) Scan() (map[string]float64, float64) {
	scores := make(map[string]float64)
	ei.lock.RLock()
	for _, item := range ei.Items {
		scores[item.MetricHash] = item.Score
	}
	ei.lock.RUnlock()
	var total float64
	for _, value := range scores {
		total += value
	}
	return scores, total
}

type (
	eventGroup struct {
		Groups map[string]*eventItems `json:"groups,omitempty"`
		lock   sync.RWMutex
	}
)

func (eg *eventGroup) Add(item *eventItem) {
	eg.lock.RLock()
	items := eg.Groups[item.GroupHash]
	eg.lock.RUnlock()
	if items == nil {
		items = &eventItems{
			Items: make(map[string]*eventItem),
		}
		eg.lock.Lock()
		eg.Groups[item.GroupHash] = items
		eg.lock.Unlock()
	}
	items.Add(item)
}

func (eg *eventGroup) Remove(groupHash, metricValueHash string) bool {
	eg.lock.RLock()
	items := eg.Groups[groupHash]
	eg.lock.RUnlock()
	if items != nil {
		if items.Remove(metricValueHash) {
			if len(items.Items) == 0 {
				eg.lock.Lock()
				delete(eg.Groups, groupHash)
				eg.lock.Unlock()
			}
			return true
		}
	}
	return false
}

type scoreForItems struct {
	Total  float64            `json:"total,omitempty"`
	Scores map[string]float64 `json:"scores,omitempty"`
}

func (eg *eventGroup) Scan() map[string]scoreForItems {
	eg.lock.RLock()
	defer eg.lock.RUnlock()
	result := make(map[string]scoreForItems, len(eg.Groups))
	for key, items := range eg.Groups {
		scores, total := items.Scan()
		result[key] = scoreForItems{
			Total:  total,
			Scores: scores,
		}
	}
	return result
}

var (
	cachedEvents     = make(map[int]*eventGroup)
	cachedEventsLock sync.RWMutex
)

func setEventItem(item *eventItem) {
	cachedEventsLock.RLock()
	group := cachedEvents[item.MultiStrategyID]
	cachedEventsLock.RUnlock()
	if group == nil {
		group = &eventGroup{
			Groups: make(map[string]*eventItems),
		}
		cachedEventsLock.Lock()
		cachedEvents[item.MultiStrategyID] = group
		cachedEventsLock.Unlock()
	}
	group.Add(item)
	judgeHitCount.Incr(1)
}

func removeEventItem(mid int, groupHash, metricValueHash string) bool {
	cachedEventsLock.RLock()
	group := cachedEvents[mid]
	cachedEventsLock.RUnlock()
	if group != nil {
		if group.Remove(groupHash, metricValueHash) {
			judgeRemoveCount.Incr(1)
			return true
		}
	}
	return false
}

var (
	judgeGroupCount      = expvar.NewBase("judge.scan_group")
	judgeHitCount        = expvar.NewDiff("judge.hit")
	judgeRemoveCount     = expvar.NewDiff("judge.remove")
	judgeScoreAlarmCount = expvar.NewDiff("judge.score")
)

func init() {
	expvar.Register(judgeGroupCount, judgeHitCount, judgeRemoveCount, judgeScoreAlarmCount)
}

func scanItems() {
	cachedEventsLock.RLock()
	var count int
	for muID, group := range cachedEvents {
		unit := GetUnit(muID)
		if unit == nil {
			log.Warn("scan-nil-unit", "muid", muID)
			continue
		}
		scoreResults := group.Scan()
		for hash, result := range scoreResults {
			ok := unit.CheckScore(result.Total)
			log.Debug("check-score", "ok", ok, "hash", hash, "result", result)
			if ok {
				judgeScoreAlarmCount.Incr(1)
			}
		}
		count += len(group.Groups)
	}
	log.Info("scan-events", "groups", count)
	judgeGroupCount.Set(int64(count))
	dumpCachedEvents()
	cachedEventsLock.RUnlock()
}

// ScanForEvents scans cached events to trigger combined-event
func ScanForEvents(interval time.Duration) {
	readCachedEvents()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		go scanItems()
	}
}

// AllCachedEvents returns copy of cached events
func AllCachedEvents() map[int]*eventGroup {
	cp := make(map[int]*eventGroup)
	cachedEventsLock.RLock()
	defer cachedEventsLock.RUnlock()
	for key, group := range cachedEvents {
		cp[key] = group
	}
	return cp
}

var (
	cachedEventsFile string
)

// SetCachedEventsFile sets file to dump cached events
func SetCachedEventsFile(file string) {
	cachedEventsFile = file
	log.Info("set-dump-file", "file", file)
}

func dumpCachedEvents() {
	if cachedEventsFile == "" {
		return
	}
	b, _ := json.Marshal(cachedEvents)
	ioutil.WriteFile(cachedEventsFile, b, os.ModePerm)
}

func readCachedEvents() {
	if cachedEventsFile == "" {
		return
	}
	b, err := ioutil.ReadFile(cachedEventsFile)
	if err != nil {
		log.Warn("read-dump-error", "error", err, "file", cachedEventsFile)
		return
	}
	cachedEventsLock.Lock()
	if err = json.Unmarshal(b, &cachedEvents); err != nil {
		log.Warn("read-dump-decode-error", "error", err, "file", cachedEventsFile)
	}
	cachedEventsLock.Unlock()
}

package multijudge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/extralib/configapi"
)

type (
	// ScoreItem is item for one score
	ScoreItem struct {
		Metric          *models.Metric `json:"metric,omitempty"`
		MultiStrategyID int            `json:"multi_strategy_id,omitempty"`
		GroupHash       string         `json:"group_hash,omitempty"`
		MetricHash      string         `json:"metric_hash,omitempty"`
		MetricValueHash string         `json:"-"`
		LeftValue       float64        `json:"left_value,omitempty"`
		Score           float64        `json:"score,omitempty"`
		strategy        *models.Strategy
	}
	// ScoreItems is score items group
	ScoreItems struct {
		Items      map[string]*ScoreItem       `json:"items,omitempty"`
		Startegies map[string]*models.Strategy `json:"startegies,omitempty"`
		lock       sync.RWMutex
	}
)

// Add adds item to score items
func (ei *ScoreItems) Add(item *ScoreItem) {
	ei.lock.Lock()
	ei.Items[item.MetricValueHash] = item
	ei.Startegies[item.Metric.Name] = item.strategy
	ei.lock.Unlock()
}

// Remove removes one value in items
func (ei *ScoreItems) Remove(metricValueHash string) bool {
	ei.lock.Lock()
	_, ok := ei.Items[metricValueHash]
	if ok {
		delete(ei.Items, metricValueHash)
	}
	ei.lock.Unlock()
	return ok
}

// ScoreResult is result of score items
type ScoreResult struct {
	Total      float64
	Scores     map[string]float64
	LeftValues map[string]float64
}

// Scan scans the items to generate score result
func (ei *ScoreItems) Scan() ScoreResult {
	scores := make(map[string]float64)
	leftValues := make(map[string]float64)
	ei.lock.RLock()
	for _, item := range ei.Items {
		if item.Score > scores[item.MetricHash] {
			scores[item.MetricHash] = item.Score
			leftValues[item.MetricHash] = item.LeftValue
		}
	}
	ei.lock.RUnlock()
	var total float64
	for _, value := range scores {
		total += value
	}
	return ScoreResult{
		Total:      total,
		Scores:     scores,
		LeftValues: leftValues,
	}
}

type (
	// ScoreGroup is group of items for one multi-strategy
	ScoreGroup struct {
		Groups map[string]*ScoreItems `json:"groups,omitempty"`
		lock   sync.RWMutex
	}
)

// Add adds item to group
func (eg *ScoreGroup) Add(item *ScoreItem) {
	eg.lock.RLock()
	items := eg.Groups[item.GroupHash]
	eg.lock.RUnlock()
	if items == nil {
		items = &ScoreItems{
			Items:      make(map[string]*ScoreItem),
			Startegies: make(map[string]*models.Strategy),
		}
		eg.lock.Lock()
		eg.Groups[item.GroupHash] = items
		eg.lock.Unlock()
	}
	items.Add(item)
}

// Remove removes item by hash in group
func (eg *ScoreGroup) Remove(groupHash, metricValueHash string) bool {
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

// Scan scans the group to gets all results
func (eg *ScoreGroup) Scan() map[string]ScoreResult {
	eg.lock.RLock()
	defer eg.lock.RUnlock()
	result := make(map[string]ScoreResult, len(eg.Groups))
	for key, items := range eg.Groups {
		result[key] = items.Scan()
	}
	return result
}

var (
	cachedEvents     = make(map[int]*ScoreGroup)
	cachedEventsLock sync.RWMutex
)

func setEventItem(item *ScoreItem) {
	cachedEventsLock.RLock()
	group := cachedEvents[item.MultiStrategyID]
	cachedEventsLock.RUnlock()
	if group == nil {
		group = &ScoreGroup{
			Groups: make(map[string]*ScoreItems),
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

func scanItems() []*models.Event {
	cachedEventsLock.Lock()
	var count int
	var nowUnix = time.Now().Unix()
	var events []*models.Event
	for muID, group := range cachedEvents {
		unit := GetUnit(muID)
		if unit == nil {
			log.Warn("remove-nil-unit", "muid", muID)
			delete(cachedEvents, muID)
			continue
		}
		scoreResults := group.Scan()
		for hash, result := range scoreResults {
			trigger := unit.CheckScore(result.Total)
			event := &models.Event{
				Time:       nowUnix,
				Status:     models.EventOk,
				Expression: muID,
				ID:         fmt.Sprintf("e_%d_%s", muID, utils.MD5HashString(hash)),
				LeftValue:  result.Total,
				Fields: map[string]interface{}{
					"total":     result.Total,
					"score":     result.Scores,
					"leftvalue": result.LeftValues,
				},
			}
			if trigger {
				judgeScoreAlarmCount.Incr(1)
				event.Status = models.EventProblem
			}
			events = append(events, event)
			log.Debug("check-event", "status", event.Status, "event", event)
		}
		count += len(group.Groups)
	}
	log.Info("scan", "groups", count, "events", len(events))
	judgeGroupCount.Set(int64(count))
	dumpCachedEvents()
	cachedEventsLock.Unlock()
	return events
}

// ScanForEvents scans cached events to trigger combined-event
func ScanForEvents(interval time.Duration) {
	readCachedEvents()
	var expHash string
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		exprList, hash := configapi.CheckExpressionsCache(expHash)
		if hash != expHash {
			expHash = hash
			SetExpressions(exprList)
			log.Info("reload-expressions", "expr", len(exprList))
		}
		events := scanItems()
		if len(events) > 0 {

		}
	}
}

// AllCachedEvents returns copy of cached events
func AllCachedEvents() map[int]*ScoreGroup {
	cp := make(map[int]*ScoreGroup)
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

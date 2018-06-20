package multijudge

import (
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

type (
	eventItem struct {
		Metric          *models.Metric `json:"metric,omitempty"`
		MultiStrategyID int            `json:"multi_strategy_id,omitempty"`
		GroupHash       string         `json:"group_hash,omitempty"`
		MetricHash      string         `json:"metric_hash,omitempty"`
		MetricValueHash string         `json:"-"`
		LeftValue       float64        `json:"left_value,omitempty"`
		Score           float64        `json:"score,omitempty"`
	}
	eventItems struct {
		Items map[string]*eventItem
		lock  sync.RWMutex
	}
	eventGroup struct {
		Groups map[string]*eventItems
		lock   sync.RWMutex
	}
)

func (ei *eventItems) Add(item *eventItem) {
	ei.lock.Lock()
	ei.Items[item.MetricValueHash] = item
	ei.lock.Unlock()
}

func (ei *eventItems) Remove(metricValueHash string) bool {
	ei.lock.Lock()
	delete(ei.Items, metricValueHash)
	ei.lock.Unlock()
	return true
}

func (ei *eventItems) Scan() {
	scores := make(map[string]float64)
	ei.lock.RLock()
	for _, item := range ei.Items {
		scores[item.MetricHash] = item.Score
	}
	ei.lock.RUnlock()
}

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
		return items.Remove(metricValueHash)
	}
	return false
}

func (eg *eventGroup) Scan() {
	eg.lock.RLock()
	defer eg.lock.RUnlock()
	for _, items := range eg.Groups {
		items.Scan()
	}
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
}

func removeEventItem(mid int, groupHash, metricValueHash string) bool {
	cachedEventsLock.RLock()
	group := cachedEvents[mid]
	cachedEventsLock.RUnlock()
	if group != nil {
		return group.Remove(groupHash, metricValueHash)
	}
	return false
}

func scanItems() {
	cachedEventsLock.RLock()
	defer cachedEventsLock.RUnlock()
	var count int
	for _, group := range cachedEvents {
		group.Scan()
		count += len(group.Groups)
	}
	log.Info("scan-events", "groups", count)
}

func ScanForEvents(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		go scanItems()
	}
}

func AllCachedEvents() map[int]*eventGroup {
	cp := make(map[int]*eventGroup)
	cachedEventsLock.RLock()
	defer cachedEventsLock.RUnlock()
	for key, group := range cachedEvents {
		cp[key] = group
	}
	return cp
}

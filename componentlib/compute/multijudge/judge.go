package multijudge

import (
	"sync"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	units        = make(map[int]*MultiUnit)
	unitsLock    sync.RWMutex
	unitsAccepts map[string][]int
)

// SetStrategies sets multi-strategies
func SetStrategies(mstList map[int]*MultiStrategy) {
	unitsLock.Lock()
	accepts := make(map[string][]int)
	for id, ss := range mstList {
		mu, err := NewMultiUnit(id, ss)
		if err != nil || mu == nil {
			log.Warn("new-munit-error", "id", id, "error", err)
			continue
		}
		units[id] = mu
		for _, m := range mu.AcceptedMetrics() {
			accepts[m] = append(accepts[m], id)
		}
	}
	unitsAccepts = accepts
	unitsLock.Unlock()
}

// GetUnit gets unit by id
func GetUnit(muID int) *MultiUnit {
	unitsLock.RLock()
	defer unitsLock.RUnlock()
	return units[muID]
}

// JudgeSingle judges one metric
func JudgeSingle(metric *models.Metric) {
	unitsLock.RLock()
	ids := unitsAccepts[metric.Name]
	if len(ids) > 0 {
		for _, id := range ids {
			mu := units[id]
			if mu == nil {
				continue
			}
			subKey, ok := mu.Accept(metric)
			if !ok {
				continue
			}
			mu.Check(subKey, metric)
		}
	}
	unitsLock.RUnlock()
}

// Judge judges some metrics
func Judge(metrics []*models.Metric) {
	for _, metric := range metrics {
		JudgeSingle(metric)
	}
}

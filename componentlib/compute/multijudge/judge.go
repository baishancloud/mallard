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

func SetStrategies(mstList map[int]*MultiStrategy) {
	unitsLock.Lock()
	accepts := make(map[string][]int)
	for id, ss := range mstList {
		mu := NewMultiUnit(id, ss)
		if mu == nil {
			log.Warn("new-munit-nil", "id", id)
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

func Judge(metrics []*models.Metric) {
	for _, metric := range metrics {
		JudgeSingle(metric)
	}
}

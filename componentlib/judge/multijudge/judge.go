package multijudge

import (
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	units        = make(map[int]*ExprUnit)
	unitsLock    sync.RWMutex
	unitsAccepts map[string][]int
)

// SetExpressions sets expression to generate unit
func SetExpressions(exprList map[int]*models.Expression) {
	nowUnix := time.Now().Unix()
	unitsLock.Lock()

	for id, expr := range exprList {
		unit := units[id]
		if unit == nil {
			unit, err := NewExprUnit(id, expr)
			if err != nil {
				log.Warn("new-exprunit-error", "id", id, "error", err)
				continue
			}
			units[id] = unit
		}
		if unit != nil {
			unit.lastTouchTime = nowUnix
		}
	}

	accepts := make(map[string][]int)
	for id, unit := range units {
		if unit.lastTouchTime > 0 && unit.lastTouchTime != nowUnix {
			log.Warn("exprunit-close", "id", id)
			delete(units, id)
			continue
		}
		for _, metric := range unit.AcceptedMetrics() {
			accepts[metric] = append(accepts[metric], id)
		}
	}

	unitsAccepts = accepts
	unitsLock.Unlock()
}

// GetUnit gets unit by id
func GetUnit(muID int) *ExprUnit {
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
			unit := units[id]
			if unit == nil {
				continue
			}
			subKey, ok := unit.Accept(metric)
			if !ok {
				continue
			}
			unit.Check(subKey, metric)
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

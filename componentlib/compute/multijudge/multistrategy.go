package multijudge

import (
	"fmt"
	"sync"

	"github.com/baishancloud/mallard/componentlib/agent/judger"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log = zaplog.Zap("mjudge")
)

// MultiStrategy is some strategies to judge one event as score comparison
type MultiStrategy struct {
	Strategies      []*models.Strategy `json:"strategies,omitempty"`
	ScoreOperator   string             `json:"score_operator,omitempty"`
	ScoreRightValue float64            `json:"score_right_value,omitempty"`
}

// MultiUnit is unit to handle one multi strategy
type MultiUnit struct {
	units map[string]*judger.StrategyUnit
	lock  sync.RWMutex

	metrics    []string
	id         int
	compareFn  judger.CompareFunc
	rightValue float64
}

// NewMultiUnit creates multi-strategy unit
func NewMultiUnit(id int, mst *MultiStrategy) *MultiUnit {
	mu := &MultiUnit{
		units:      make(map[string]*judger.StrategyUnit, len(mst.Strategies)),
		id:         id,
		compareFn:  judger.NewCompareFunc(mst.ScoreOperator),
		rightValue: mst.ScoreRightValue,
	}
	for _, st := range mst.Strategies {
		unit, err := judger.NewUnit(st)
		if err != nil {
			log.Warn("new-unit-error", "error", err, "st", st)
			continue
		}
		key := st.Metric + "-" + st.TagString
		mu.units[key] = unit
		mu.metrics = append(mu.metrics, st.Metric)
	}
	mu.metrics = utils.StringSliceUnique(mu.metrics)
	return mu
}

// Accept checks metric to accepting
// returns sub-strategy key and accepted status
func (mu *MultiUnit) Accept(metric *models.Metric) (string, bool) {
	mu.lock.RLock()
	for key, unit := range mu.units {
		if unit.Accept(metric) {
			mu.lock.RUnlock()
			return key, true
		}
	}
	mu.lock.RUnlock()
	return "", false
}

// AcceptedMetrics returns accepting metrics from the unit
func (mu *MultiUnit) AcceptedMetrics() []string {
	return mu.metrics
}

// Check checks value with sub-key
func (mu *MultiUnit) Check(key string, metric *models.Metric) {
	mu.lock.RLock()
	unit := mu.units[key]
	mu.lock.RUnlock()
	if unit == nil {
		return
	}
	groups := unit.GroupBy()
	fullTags := metric.FullTags()
	var groupHash string
	for _, keyword := range groups {
		tag := fullTags[keyword]
		if tag == "" {
			tag = "|"
		}
		groupHash += tag + "-"
	}
	leftValue, status, err := unit.Check(metric, groupHash)
	if err != nil {
		log.Warn("check-error", "metric", metric.Name, "id", mu.id, "err", err)
		return
	}
	metricHash := fmt.Sprintf("%d~%s~%s", mu.id, groupHash, metric.Name)
	metricValueHash := metricHash + "~" + metric.Hash()
	if status == models.EventProblem {
		item := &eventItem{
			Metric:          metric,
			MultiStrategyID: mu.id,
			GroupHash:       groupHash,
			MetricHash:      metricHash,
			MetricValueHash: metricValueHash,
			LeftValue:       leftValue,
			Score:           unit.Score(),
			Strategy:        unit.GetStrategy(),
		}
		setEventItem(item)
		log.Debug("set-event", "mhash", metricHash, "vhash", metricValueHash, "left", leftValue)
	} else {
		if removeEventItem(mu.id, groupHash, metricValueHash) {
			log.Debug("remove-event", "mhash", metricHash, "vhash", metricValueHash, "left", leftValue)
		}
		// tryRemoveProblem(problemKey)
	}
}

// CheckScore checks score with compare function
func (mu *MultiUnit) CheckScore(leftValue float64) bool {
	if mu.compareFn == nil {
		return false
	}
	return mu.compareFn(leftValue, mu.rightValue)
}

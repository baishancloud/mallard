package judger

import (
	"errors"
	"fmt"
	"sync"

	"github.com/baishancloud/mallard/corelib/models"
)

// StrategyUnit is a unit to check strategy in judge
type StrategyUnit struct {
	st *models.Strategy
	op Operator

	dataQueue map[string][]*models.EventValue
	lock      sync.RWMutex
}

// NewUnit return new judge unit with strategy
func NewUnit(st *models.Strategy) (*StrategyUnit, error) {
	su := &StrategyUnit{
		dataQueue: make(map[string][]*models.EventValue),
	}
	return su, su.SetStrategy(st)
}

// ID is unit id, same to strategy id
func (s *StrategyUnit) ID() int {
	if s.st == nil {
		return 0
	}
	return s.st.ID
}

// SetStrategy sets strategy data
func (s *StrategyUnit) SetStrategy(st *models.Strategy) error {
	if st == nil {
		return errors.New("nil")
	}
	op, err := FromStrategy(st)
	if err != nil {
		return fmt.Errorf("strategy-%d-parse-error-%s", st.ID, err.Error())
	}
	s.st = st
	s.op = op
	return nil
}

// GetStrategy gets strategy data in unit
func (s *StrategyUnit) GetStrategy() *models.Strategy {
	return s.st
}

func (s *StrategyUnit) genQueue(metric *models.Metric, rawLeftValue float64, customHash string) ([]*models.EventValue, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	hash := customHash
	if hash == "" {
		hash = metric.Hash()
	}
	queue := s.dataQueue[hash]
	historyData := &models.EventValue{Value: rawLeftValue, Time: metric.Time}
	if len(queue) > 0 {
		if historyData.Time < queue[0].Time { // accept latest data
			return nil, false
		}
	}
	queue = append([]*models.EventValue{historyData}, queue...)
	if len(queue) < s.op.Limit() {
		s.dataQueue[hash] = queue
		return nil, false
	}
	queue = queue[:s.op.Limit()]
	s.dataQueue[hash] = queue
	return queue, true
}

// Check check metric value to bool result, if problem, return false
func (s *StrategyUnit) Check(metric *models.Metric, customHash string) (float64, models.EventStatus, error) {
	if s.st == nil {
		return 0, models.EventIgnore, nil
	}
	rawLeftValue, err := s.op.Transform(metric)
	if err != nil {
		return 0, models.EventIgnore, err
	}
	queue, ok := s.genQueue(metric, rawLeftValue, customHash)
	if !ok {
		return 0, models.EventIgnore, nil
	}
	leftValue, ok, err := s.op.Trigger(queue)
	if err != nil {
		return 0, models.EventOk, err
	}
	if ok {
		return leftValue, models.EventProblem, nil
	}
	return leftValue, models.EventOk, nil
}

// Accept check metric name is suited for this unit
func (s *StrategyUnit) Accept(metric *models.Metric) bool {
	if s.st == nil || s.op == nil {
		return false
	}
	if s.st.Metric != metric.Name {
		return false
	}
	tagFilters := s.op.Tags()
	if len(tagFilters) > 0 {
		fullTags := metric.FullTags()
		if len(fullTags) == 0 {
			return false
		}
		for key, tagFunc := range tagFilters {
			v := fullTags[key]
			if !tagFunc(v) {
				return false
			}
		}
	}
	return true
}

// History return history data in this unit with metric hash key
func (s *StrategyUnit) History(hash string) []*models.EventValue {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.dataQueue[hash]
}

// Operator return operator in the strategy unit
func (s *StrategyUnit) Operator() Operator {
	return s.op
}

var (
	defaultUnitGroups = []string{"endpoint"}
)

// GroupBy returns unit custom group by string
func (s *StrategyUnit) GroupBy() []string {
	if s.st == nil {
		return nil
	}
	ss := s.st.GroupBy()
	if len(ss) == 0 {
		return defaultUnitGroups
	}
	return ss
}

// Score gets score of the strategy in unit
func (s *StrategyUnit) Score() float64 {
	if s.st == nil {
		return 0
	}
	return s.st.Score
}

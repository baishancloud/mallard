package multijudge

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	ID              int      `json:"id,omitempty"`
	StrategiesRules []string `json:"rules,omitempty"`
	ScoreExpr       string   `json:"score_expr,omitempty"`
	Note            string   `json:"note,omitempty"`
	Priority        int      `json:"priority,omitempty"`
	MaxStep         int      `json:"max_step,omitempty"`
	Step            int      `json:"step,omitempty"`
	Nodata          int      `json:"nodata,omitempty"`
	RecoverNotify   int      `json:"recover_notify,omitempty"`
	Status          int      `json:"status,omitempty"`

	strategies      []*models.Strategy
	scoreOperator   string
	scoreRightValue float64
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

var (
	// ErrorRulesEmpty means rules are empty
	ErrorRulesEmpty = errors.New("rules-empty")
	// ErrorRuleInvalidLength means rule segments are not correct length
	ErrorRuleInvalidLength = errors.New("rule-invalid-length")
	// ErrorUnknownScoreExpr means can not parse score expression
	ErrorUnknownScoreExpr = errors.New("unknown-score-expr")
)

func parseRulesToStrategy(rules []string) ([]*models.Strategy, error) {
	if len(rules) == 0 {
		return nil, ErrorRulesEmpty
	}
	ss := make([]*models.Strategy, 0, len(rules))
	for _, rule := range rules {
		st, err := parseOneRule(rule)
		if err != nil {
			return nil, err
		}
		if st != nil {
			ss = append(ss, st)
		}
	}
	return ss, nil
}

func parseOneRule(rule string) (*models.Strategy, error) {
	segments := strings.Split(rule, ";")
	if len(segments) < 6 {
		return nil, ErrorRuleInvalidLength
	}
	st := &models.Strategy{
		Metric:         segments[0],
		FieldTransform: segments[1],
		TagString:      segments[3],
		GroupBys:       strings.Split(segments[4], ","),
	}
	var err error
	st.Score, err = strconv.ParseFloat(segments[5], 64)
	if err != nil {
		return nil, err
	}
	idx := strings.Index(segments[2], ")")
	if idx < 0 {
		return nil, ErrorRuleInvalidLength
	}
	st.Func = segments[2][:idx+1]
	st.Operator, st.RightValue, err = parseScoreOperator(segments[2][idx+1:])
	if err != nil {
		return nil, err
	}
	// log.Debug("parse-strategy", "st", st)
	return st, nil
}

func parseScoreOperator(opt string) (string, float64, error) {
	if strings.HasPrefix(opt, ">=") {
		value, err := strconv.ParseFloat(strings.TrimPrefix(opt, ">="), 64)
		return ">=", value, err
	}
	if strings.HasPrefix(opt, "<=") {
		value, err := strconv.ParseFloat(strings.TrimPrefix(opt, "<="), 64)
		return "<=", value, err
	}
	if strings.HasPrefix(opt, "==") {
		value, err := strconv.ParseFloat(strings.TrimPrefix(opt, "=="), 64)
		return "==", value, err
	}
	if strings.HasPrefix(opt, ">") {
		value, err := strconv.ParseFloat(strings.TrimPrefix(opt, ">"), 64)
		return ">", value, err
	}
	if strings.HasPrefix(opt, "<") {
		value, err := strconv.ParseFloat(strings.TrimPrefix(opt, "<"), 64)
		return "<", value, err
	}
	if strings.HasPrefix(opt, "=") {
		value, err := strconv.ParseFloat(strings.TrimPrefix(opt, "="), 64)
		return "==", value, err
	}
	return "", 0, ErrorUnknownScoreExpr
}

// NewMultiUnit creates multi-strategy unit
func NewMultiUnit(id int, mst *MultiStrategy) (*MultiUnit, error) {
	op, rightValue, err := parseScoreOperator(mst.ScoreExpr)
	if err != nil {
		return nil, err
	}
	mu := &MultiUnit{
		units:      make(map[string]*judger.StrategyUnit, len(mst.StrategiesRules)),
		id:         id,
		compareFn:  judger.NewCompareFunc(op),
		rightValue: rightValue,
	}
	if mst.strategies, err = parseRulesToStrategy(mst.StrategiesRules); err != nil {
		return nil, err
	}
	for _, st := range mst.strategies {
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
	return mu, nil
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
	groupHash = strings.TrimRight(groupHash, "-")
	groupHash = fmt.Sprintf("%d~%s", mu.id, groupHash)
	leftValue, status, err := unit.Check(metric, groupHash)
	if err != nil {
		log.Warn("check-error", "metric", metric.Name, "id", mu.id, "err", err)
		return
	}
	metricHash := fmt.Sprintf("%s~%s", groupHash, metric.Name)
	metricValueHash := metric.Name + "~" + metric.Hash()
	if status == models.EventProblem {
		item := &ScoreItem{
			Metric:          metric,
			MultiStrategyID: mu.id,
			GroupHash:       groupHash,
			MetricHash:      metricHash,
			MetricValueHash: metricValueHash,
			LeftValue:       leftValue,
			Score:           unit.Score(),
			strategy:        unit.GetStrategy(),
		}
		setEventItem(item)
		log.Debug("set", "mhash", metricHash, "vhash", metricValueHash, "left", leftValue)
	} else {
		if removeEventItem(mu.id, groupHash, metricValueHash) {
			log.Debug("remove", "mhash", metricHash, "vhash", metricValueHash, "left", leftValue)
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

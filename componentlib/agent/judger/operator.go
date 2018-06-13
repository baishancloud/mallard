package judger

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

// Operator is a judge operator with one strategy or expression
type Operator interface {
	Limit() int
	Transform(metric *models.Metric) (float64, error)
	Trigger(historyList []*models.EventValue) (float64, bool, error)
	Tags() map[string]TagFilterFunc
	Base() *StrategyBase
}

// FromStrategy return operator with strategy
func FromStrategy(s *models.Strategy) (Operator, error) {
	if strings.HasPrefix(s.FieldTransform, "select(") {
		return NewSelectFromStrategy(s)
	}
	if strings.HasPrefix(s.FieldTransform, "rangeselect(") {
		return NewRangeSelectFromStrategy(s)
	}
	return nil, errors.New("unknown transform")
}

var (
	calReplacer = strings.NewReplacer(")", "", "(#", "|", ",", "|")
)

// StrategyBase is base info after strategy parsed
type StrategyBase struct {
	Metric     string
	CalType    string
	Limit      int
	Tags       map[string]TagFilterFunc
	Args       []interface{}
	Field      string
	OpSign     string
	RightValue float64
}

// NewStrategyBase parse strategy model to base info
func NewStrategyBase(st *models.Strategy) (*StrategyBase, error) {
	callStrings := strings.Split(calReplacer.Replace(st.Func), "|")
	if len(callStrings) == 3 {
		return parseStrategyThree(st, callStrings)
	}
	if len(callStrings) != 2 {
		return nil, errors.New("judge.operator : strategy func is invalid")
	}
	stb := &StrategyBase{}
	limit, err := strconv.Atoi(callStrings[1])
	if err != nil {
		return nil, err
	}
	stb.Limit = limit
	stb.Tags = make(map[string]TagFilterFunc)
	rawTags := make(map[string]string)
	rawTags, err = models.ExtractTags(st.TagString)
	if err != nil {
		return nil, err
	}
	for k, v := range rawTags {
		key, fn := NewTagFilterFunc(k, v)
		stb.Tags[key] = fn
	}
	stb.CalType = callStrings[0]
	stb.Metric = st.Metric
	stb.Limit = getCalculateLimit(stb.Limit, stb.CalType)
	stb.OpSign = st.Operator
	stb.RightValue = st.RightValue
	return stb, nil
}

func parseStrategyThree(st *models.Strategy, callStrings []string) (*StrategyBase, error) {
	stb := &StrategyBase{}
	limit, err := strconv.Atoi(callStrings[1])
	if err != nil {
		return nil, err
	}
	stb.Limit = limit

	threeArg, err := strconv.Atoi(callStrings[2])
	if err != nil {
		return nil, err
	}
	stb.Args = []interface{}{threeArg}

	stb.Tags = make(map[string]TagFilterFunc)
	rawTags := make(map[string]string)
	rawTags, err = models.ExtractTags(st.TagString)
	if err != nil {
		return nil, err
	}
	for k, v := range rawTags {
		key, fn := NewTagFilterFunc(k, v)
		stb.Tags[key] = fn
	}
	stb.CalType = callStrings[0]
	stb.Metric = st.Metric
	stb.Limit = getCalculateLimit(stb.Limit, stb.CalType)
	stb.OpSign = st.Operator
	stb.RightValue = st.RightValue
	return stb, nil
}

var selectReplacer = strings.NewReplacer("(", "|", ")", "")

// Select is operator for select()
type Select struct {
	base        *StrategyBase
	compareFn   CompareFunc
	calculateFn CalculateFunc
}

// NewSelectFromStrategy return select operator with strategy
func NewSelectFromStrategy(st *models.Strategy) (*Select, error) {
	var (
		s   Select
		err error
	)
	s.base, err = NewStrategyBase(st)
	if err != nil {
		return nil, err
	}
	s.calculateFn = NewCalculateFunc(s.base.CalType)
	if s.calculateFn == nil {
		return nil, err
	}
	s.compareFn = NewCompareFunc(st.Operator)
	if s.compareFn == nil {
		return nil, err
	}
	fieldS := strings.Split(selectReplacer.Replace(st.FieldTransform), "|")
	if len(fieldS) != 2 {
		return nil, err
	}
	s.base.Field = strings.TrimSpace(fieldS[1])
	return &s, nil
}

// FieldMissingError is type to define missing field error
type FieldMissingError struct {
	error
}

// Limit return data queue size
func (s *Select) Limit() int {
	return s.base.Limit
}

// Transform return field value from metric by operator
func (s *Select) Transform(metric *models.Metric) (float64, error) {
	if s.base.Field == "value" {
		return metric.Value, nil
	}
	fieldv, ok := metric.Fields[s.base.Field]
	if !ok {
		return 0, FieldMissingError{error: fmt.Errorf("field '%s' is not found", s.base.Field)}
	}
	return utils.ToFloat64(fieldv)
}

// Trigger check history data with strategy
func (s *Select) Trigger(historyList []*models.EventValue) (float64, bool, error) {
	return s.calculateFn(historyList, s.base.RightValue, s.compareFn, s.base.Args...)
}

// Tags return strategy tags condition
func (s *Select) Tags() map[string]TagFilterFunc {
	if s.base == nil {
		return nil
	}
	return s.base.Tags
}

// Base return strategy base info
func (s *Select) Base() *StrategyBase {
	return s.base
}

var rangeSelectReplacer = strings.NewReplacer("(", "|", ")", "", ",", "|")

// RangeInfo is params for range select
type RangeInfo struct {
	RangeField string
	MinValue   float64
	MaxValue   float64
	RangeValue float64
}

// RangeSelect is operator for rangeselect()
type RangeSelect struct {
	base *StrategyBase

	compareFn   CompareFunc
	calculateFn CalculateFunc
	rangeInfo   RangeInfo
}

// NewRangeSelectFromStrategy return rangeselect operator
func NewRangeSelectFromStrategy(st *models.Strategy) (*RangeSelect, error) {
	var (
		s   RangeSelect
		err error
	)
	s.base, err = NewStrategyBase(st)
	if err != nil {
		return nil, err
	}
	s.calculateFn = NewCalculateFunc(s.base.CalType)
	if s.calculateFn == nil {
		return nil, err
	}
	s.compareFn = NewCompareFunc(st.Operator)
	if s.compareFn == nil {
		return nil, err
	}
	fieldS := strings.Split(rangeSelectReplacer.Replace(st.FieldTransform), "|")
	if len(fieldS) != 6 {
		return nil, errors.New("rangeselect(args) count wrong")
	}
	s.base.Field = strings.TrimSpace(fieldS[1])
	if s.rangeInfo.MinValue, err = strconv.ParseFloat(fieldS[2], 64); err != nil {
		return nil, err
	}
	if s.rangeInfo.MaxValue, err = strconv.ParseFloat(fieldS[3], 64); err != nil {
		return nil, err
	}
	s.rangeInfo.RangeField = strings.TrimSpace(fieldS[4])
	if s.rangeInfo.RangeValue, err = strconv.ParseFloat(fieldS[5], 64); err != nil {
		return nil, err
	}
	return &s, nil
}

// Limit return data queue size limit
func (s *RangeSelect) Limit() int {
	return s.base.Limit
}

// Transform return left value from metric with this operator
// get field value by field,
// if value >= min && value <= max, return range field value
func (s *RangeSelect) Transform(metric *models.Metric) (float64, error) {
	fieldv, ok := metric.Fields[s.base.Field]
	if !ok {
		if s.base.Field == "value" {
			fieldv = metric.Value
		} else {
			return 0, fmt.Errorf("field '%s' is not found", s.base.Field)
		}
	}
	fieldFloat, err := utils.ToFloat64(fieldv)

	if err != nil {
		return 0, err
	}
	if fieldFloat >= s.rangeInfo.MinValue && fieldFloat <= s.rangeInfo.MaxValue {
		fieldv2, ok := metric.Fields[s.rangeInfo.RangeField]
		if !ok {
			if s.rangeInfo.RangeField == "value" {
				fieldv2 = metric.Value
			} else {
				return 0, fmt.Errorf("field '%s' is not found", s.rangeInfo.RangeField)
			}
		}
		return utils.ToFloat64(fieldv2)
	}
	return s.rangeInfo.RangeValue, nil
}

// Trigger run strategy rule with history data list
func (s *RangeSelect) Trigger(historyList []*models.EventValue) (float64, bool, error) {
	return s.calculateFn(historyList, s.base.RightValue, s.compareFn, s.base.Args...)
}

// Tags return strategy tags condition
func (s *RangeSelect) Tags() map[string]TagFilterFunc {
	if s.base == nil {
		return nil
	}
	return s.base.Tags
}

// Base return strategy base info
func (s *RangeSelect) Base() *StrategyBase {
	return s.base
}

// RangeInfo return rangeselect params
func (s *RangeSelect) RangeInfo() RangeInfo {
	return s.rangeInfo
}

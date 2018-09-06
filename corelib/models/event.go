package models

import (
	"encoding/json"
	"strconv"
	"strings"
)

// EventValue is one data of event history
type EventValue struct {
	Value float64 `json:"value"`
	Time  int64   `json:"time"`
}

// Event is a event result after judging metric value
type Event struct {
	ID         string                 `json:"id"`
	Status     EventStatus            `json:"status"`
	Time       int64                  `json:"t"`
	Strategy   int                    `json:"st,omitempty"`
	Expression int                    `json:"exp,omitempty"`
	Endpoint   string                 `json:"ep,omitempty"`
	Step       int                    `json:"step"`
	LeftValue  float64                `json:"lv,omitempty"`
	History    []*EventValue          `json:"history,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Cycle      int                    `json:"cycle,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	CreateTime int64                  `json:"ct,omitempty"`
}

// Simple trim useless fields as empty when event keep OK too long
func (e *Event) Simple() *Event {
	return &Event{
		ID:         e.ID,
		Status:     e.Status,
		Time:       e.Time,
		Endpoint:   e.Endpoint,
		Step:       e.Step,
		CreateTime: e.CreateTime,
		Strategy:   e.Strategy,
	}
}

// JudgeUnitID gets strategy or expression id from fields or eid
func (e *Event) JudgeUnitID() int {
	if e.Strategy > 0 {
		return e.Strategy
	}
	if e.Expression > 0 {
		return e.Expression
	}
	return GetJudgeUnitID(e.ID)
}

// GetJudgeUnitID gets strategy or expression id from eid
func GetJudgeUnitID(eid string) int {
	if strings.HasPrefix(eid, "s_") || strings.HasPrefix(eid, "e_") {
		sl := strings.Split(eid, "_")
		if len(sl) < 3 {
			return 0
		}
		v, _ := strconv.Atoi(sl[1])
		return v
	}
	return 0
}

// EventStatus is status code for event
type EventStatus int

const (
	// EventOk mean event is ok
	EventOk EventStatus = 1 + iota*2
	// EventIgnore mean event is ignored because strategy or expression is invalid temporarily
	EventIgnore
	// EventSyntax mean event's strategy or expression is wrong syntax
	EventSyntax
	// EventNotEnough mean it's not enough data to check
	EventNotEnough
	// EventFieldMissing mean field from event's strategy or expression is missing
	EventFieldMissing
	// EventOutTime mean strategy or expression runs in wrong time range
	EventOutTime
	// EventProblem mean event is in problem
	EventProblem EventStatus = -1
	// EventNoData mean event is in no data case
	EventNoData EventStatus = -3
	// EventOutdated mean event is outdated
	EventOutdated EventStatus = -5
	// EventFakeOK means fake ok status, same to outdated
	EventFakeOK EventStatus = -7
	// EventMissingStrategy means event's strategy is missing
	EventMissingStrategy EventStatus = -9
	// EventClosed means event strategy is closed to the endpoint
	EventClosed EventStatus = -11
)

// String implement stringer interface
func (status EventStatus) String() string {
	switch status {
	case EventOk:
		return "OK"
	case EventProblem:
		return "PROBLEM"
	case EventIgnore:
		return "IGNORE"
	case EventSyntax:
		return "SYNTAX"
	case EventNotEnough:
		return "NOT_ENOUGH"
	case EventOutTime:
		return "OUT_TIME"
	case EventOutdated:
		return "OUTDATED"
	case EventNoData:
		return "NODATA"
	case EventFakeOK:
		return "FOK"
	case EventMissingStrategy:
		return "MISSING"
	case EventClosed:
		return "CLOSED"
	}
	return ""
}

// EventFull define event info with related strategy and expression data
type EventFull struct {
	ID          string                 `json:"id"`
	Strategy    interface{}            `json:"strategy"`
	Status      string                 `json:"status"` // OK or PROBLEM
	Endpoint    string                 `json:"endpoint"`
	LeftValue   float64                `json:"leftValue"`
	CurrentStep int                    `json:"currentStep"`
	EventTime   int64                  `json:"eventTime"`
	Judge       string                 `json:"judge"`
	PushedTags  map[string]string      `json:"pushedTags"`
	Fields      map[string]interface{} `json:"fields"`
	// Action      *AlarmAction           `json:"-"`
	// IsMaintain  bool                   `json:"-"`
}

// Priority return priority value from strategy or expression
func (ef *EventFull) Priority() int {
	if ef.Strategy != nil {
		switch rv := ef.Strategy.(type) {
		case *Strategy:
			return rv.Priority
		}
	}
	return 0
}

// FullTags return all tags with serv and endpoint data
func (e *Event) FullTags() map[string]string {
	fullTags := make(map[string]string)
	for k, v := range e.Tags {
		fullTags[k] = v
	}
	fullTags["endpoint"] = e.Endpoint
	/*
		for k, v := range splitCachegroup(e.Tags["cachegroup"]) {
			fullTags[k] = v
		}
		for k, v := range splitSertypes(e.Tags["sertypes"]) {
			fullTags[k] = v
		}*/
	return fullTags
}

func splitCachegroup(cacheGroup string) map[string]string {
	if cacheGroup == "" {
		return nil
	}
	strs := strings.Split(cacheGroup, "-")
	length := len(strs)
	if length < 6 {
		return nil
	}
	m := make(map[string]string)
	m["cachegroup_group"] = strings.Join(strs[length-2:], "-")
	m["cachegroup_group"] = strs[length-3]
	m["cachegroup_city"] = strs[length-4]
	m["cachegroup_province"] = strs[length-5]
	isp := strs[0]
	for i := 1; i < length-5; i++ {
		isp = isp + "-" + strs[i]
	}
	m["cachegroup_isp"] = isp
	return m
}

func splitSertypes(sertypes string) map[string]string {
	if sertypes == "" {
		return nil
	}
	strs := strings.Split(sertypes, "|")
	if len(strs) < 1 {
		return nil
	}
	m := make(map[string]string)
	for _, s := range strs {
		m["sertypes_"+s] = "1"
	}
	return m
}

// EventDto is todo mark for an event
type EventDto struct {
	ID             string                 `json:"id"`
	Endpoint       string                 `json:"endpoint"`
	Metric         string                 `json:"metric"`
	Counter        string                 `json:"counter"`
	FieldTransform string                 `json:"fieldTransform"`
	Func           string                 `json:"func"`
	LeftValue      string                 `json:"leftValue"`
	Operator       string                 `json:"operator"`
	RightValue     string                 `json:"rightValue"`
	Note           string                 `json:"note"`
	MaxStep        int                    `json:"maxStep"`
	CurrentStep    int                    `json:"currentStep"`
	Priority       int                    `json:"priority"`
	Status         string                 `json:"status"`
	Timestamp      int64                  `json:"timestamp"`
	ExpressionID   int                    `json:"expressionId"`
	StrategyID     int                    `json:"strategyId"`
	TemplateID     int                    `json:"templateId"`
	ActionID       int                    `json:"actionId"`
	Judge          string                 `json:"judge"`
	PushedTags     map[string]string      `json:"pushedTags"`
	Fields         map[string]interface{} `json:"fields"`
	Link           string                 `json:"link"`
	MarkTags       []string               `json:"mark_tags"`
}

// NewEventDto creates new dto from event and strategy
func NewEventDto(event *EventFull, st *Strategy) *EventDto {
	t := &EventDto{
		ID:             event.ID,
		Endpoint:       event.Endpoint,
		Metric:         st.Metric,
		FieldTransform: st.FieldTransform,
		Func:           st.Func,
		LeftValue:      friendFloat(event.LeftValue),
		Operator:       st.Operator,
		RightValue:     friendFloat(st.RightValue),
		Note:           st.Note,
		MaxStep:        st.MaxStep,
		CurrentStep:    event.CurrentStep,
		Priority:       event.Priority(),
		Status:         event.Status,
		Timestamp:      event.EventTime,
		ExpressionID:   0,
		StrategyID:     st.ID,
		TemplateID:     st.TemplateID,
		// ActionID:       event.Action.ID,
		PushedTags: event.PushedTags,
		Fields:     event.Fields,
		Judge:      event.Judge,
		Link:       "",
		MarkTags:   st.Marks(),
	}
	return t
}

// NewEventDtoByExpression creates new dto from event and expression
func NewEventDtoByExpression(event *EventFull, exp *Expression) *EventDto {
	t := &EventDto{
		ID:             event.ID,
		Endpoint:       event.Endpoint,
		Metric:         strings.Join(exp.Metrics(), ","),
		FieldTransform: "",
		Func:           "",
		LeftValue:      friendFloat(event.LeftValue),
		Operator:       exp.Operator,
		RightValue:     friendFloat(exp.RightValue),
		Note:           exp.Note,
		MaxStep:        exp.MaxStep,
		CurrentStep:    event.CurrentStep,
		Priority:       event.Priority(),
		Status:         event.Status,
		Timestamp:      event.EventTime,
		ExpressionID:   exp.ID,
		StrategyID:     0,
		TemplateID:     0,
		// ActionID:       event.Action.ID,
		PushedTags: event.PushedTags,
		Fields:     event.Fields,
		Judge:      event.Judge,
	}
	return t
}

func friendFloat(raw float64) string {
	val := strconv.FormatFloat(raw, 'f', 5, 64)
	if strings.Contains(val, ".") {
		val = strings.TrimRight(val, "0")
		val = strings.TrimRight(val, ".")
	}
	return val
}

// FillEventStrategy fills event strategy from interface to strategy
func FillEventStrategy(event *EventFull) error {
	if _, ok := event.Strategy.(*Strategy); ok { // already set, ignore
		return nil
	}
	b, _ := json.Marshal(event.Strategy)
	st := new(Strategy)
	if err := json.Unmarshal(b, st); err != nil {
		return err
	}
	event.Strategy = st
	return nil
}

// FillEventExpression fills event expression from interface to expression
func FillEventExpression(event *EventFull) error {
	if _, ok := event.Strategy.(*Expression); ok { // already set, ignore
		return nil
	}
	b, _ := json.Marshal(event.Strategy)
	exp := new(Expression)
	if err := json.Unmarshal(b, exp); err != nil {
		return err
	}
	event.Strategy = exp
	return nil
}

// EventDbData is raw data object for event from database
type EventDbData struct {
	Eid            string  `db:"eid"`
	FieldTransform string  `db:"field_transform"`
	Func           string  `db:"func"`
	Operator       string  `db:"operator"`
	RightValue     float64 `db:"right_value"`
	Endpoint       string  `db:"endpoint"`
	Status         string  `db:"status"`
	Priority       int     `db:"priority"`
	Time           int64   `db:"event_time"`
	StrategyID     int     `db:"strategy_id"`
	TemplateID     int     `db:"template_id"`
	Note           string  `db:"note"`
	Duration       int     `db:"duration"`
}

package models

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Template is template object for strategy
type Template struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"template_name"`
	ParentID int    `json:"pid,omitempty" db:"parent_id"`
	ActionID int    `json:"act" db:"action_id"`
	Creator  string `json:"-" db:"create_user"`
}

// Strategy is metriv value rule
type Strategy struct {
	ID             int      `json:"id" db:"id"`
	Metric         string   `json:"m" db:"metric"`
	FieldTransform string   `json:"tf,omitempty" db:"field_transform"` // default is "" , eq selection value.
	TagString      string   `json:"tags,omitempty" db:"tags"`
	Func           string   `json:"func" db:"func"` // e.g. max(#3) all(#3)
	Operator       string   `json:"op" db:"op"`
	RightValueStr  string   `json:"-" db:"right_value"` // e.g. < !=
	RightValue     float64  `json:"rv" db:"-"`          // critical value
	MaxStep        int      `json:"ms" db:"max_step"`
	Step           int      `json:"step" db:"step"`
	Priority       int      `json:"priority,omitempty" db:"priority"`
	Note           string   `json:"note,omitemtpy" db:"note"`
	TemplateID     int      `json:"tid" db:"template_id"`
	RunBegin       string   `json:"rb,omitempty" db:"run_begin"`
	RunEnd         string   `json:"re,omitempty" db:"run_end"`
	NoData         int      `json:"nd,omitempty" db:"no_data"`
	RecoverNotify  int      `json:"recover,omitempty" db:"recover_notify"`
	SilencesTime   int      `json:"silence,omitempty" db:"silences_time"`
	MarkTagStr     string   `json:"-" db:"mark_tags"`
	MarkTags       []string `json:"mark_tags,omitempty" db:"-"`
	Status         int      `json:"-" db:"status"`
	GroupByStr     string   `json:"-" db:"group_by"`
	GroupBys       []string `json:"group_bys,omitempty" db:"-"`
	Score          float64  `json:"score,omitempty" db:"score"`

	tags     map[string]string
	template *Template

	simple *Strategy
}

// ToSimple convert full strategy info to essentail fields
func (s *Strategy) ToSimple() *Strategy {
	if s.simple == nil {
		s.simple = &Strategy{
			ID:             s.ID,
			Metric:         s.Metric,
			FieldTransform: s.FieldTransform,
			Func:           s.Func,
			Operator:       s.Operator,
			RightValue:     s.RightValue,
			TagString:      s.TagString,
		}
	}
	return s.simple
}

// Marks return mark tags slice
func (s *Strategy) Marks() []string {
	if len(s.MarkTags) == 0 {
		if s.MarkTagStr == "" {
			return nil
		}
		s.MarkTags = strings.Split(s.MarkTagStr, ",")
	}
	return s.MarkTags
}

// GroupBy return group by string slice
func (s *Strategy) GroupBy() []string {
	if len(s.GroupBys) == 0 {
		if s.GroupByStr == "" {
			return nil
		}
		s.GroupBys = strings.Split(s.GroupByStr, ",")
		sort.Sort(sort.StringSlice(s.GroupBys))
	}
	return s.GroupBys
}

// IsInTime check t is in strategy running duration
func (s *Strategy) IsInTime(t int64) bool {
	if s.RunBegin == "" || s.RunEnd == "" {
		return true
	}
	if strings.TrimSpace(s.RunBegin) == strings.TrimSpace(s.RunEnd) {
		return false
	}
	str := time.Unix(t, 0).Format("15:04")
	if s.RunBegin < s.RunEnd {
		return str >= s.RunBegin && str <= s.RunEnd
	}
	return !(str >= s.RunEnd && str <= s.RunBegin)
}

// IsEnable checkes strategy is enabled to run
func (s *Strategy) IsEnable() bool {
	if s.Status == 0 {
		return false
	}
	if s.RunBegin != "" && s.RunEnd != "" && s.RunBegin == s.RunEnd {
		return false
	}
	return true
}

// ExtractTags extracts tag string to map
func ExtractTags(s string) (map[string]string, error) {
	if s == "" {
		return nil, nil
	}
	tags := make(map[string]string, strings.Count(s, ","))
	tagSlice := strings.Split(s, ",")
	for _, tag := range tagSlice {
		pair := strings.SplitN(tag, "=", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("bad tag %s", tag)
		}
		k := strings.TrimSpace(pair[0])
		v := strings.TrimSpace(pair[1])
		tags[k] = v

	}
	return tags, nil
}

// ExtractFields extracts field string to map
func ExtractFields(s string) (map[string]interface{}, error) {
	if s == "" {
		return nil, nil
	}
	var (
		fields = make(map[string]interface{}, strings.Count(s, ","))
	)
	fieldSlice := strings.Split(s, ",")
	for _, field := range fieldSlice {
		pair := strings.SplitN(field, "=", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("bad field %s", field)
		}
		k := strings.TrimSpace(pair[0])
		pair[1] = strings.TrimSpace(pair[1])
		vv, err := strconv.ParseFloat(pair[1], 64)
		if err != nil {
			err = nil
			vs, err := strconv.Unquote(pair[1])
			if err == nil {
				fields[k] = vs
			} else {
				fields[k] = pair[1]
			}
			continue
		}
		fields[k] = vv
		continue
	}
	return fields, nil
}

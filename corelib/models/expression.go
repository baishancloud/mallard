package models

import (
	"encoding/json"
	"strings"
)

// Expression is expression for metrics judger
type Expression struct {
	ID            int     `json:"id" db:"id"`
	Expression    string  `json:"expression" db:"expression"`
	Operator      string  `json:"operator" db:"op"`
	RightValueStr string  `json:"-" db:"right_value"` // critical value
	RightValue    float64 `json:"right_value" db:"-"` // critical value
	MaxStep       int     `json:"max_step" db:"max_step"`
	Priority      int     `json:"priority" db:"priority"`
	Note          string  `json:"note" db:"note"`

	parsedMetrics []string
}

// Metrics returns metrics from Expression rules
func (exp *Expression) Metrics() []string {
	if len(exp.parsedMetrics) == 0 {
		var rules []string
		var parsedMetrics []string
		json.Unmarshal([]byte(exp.Expression), &rules)
		if len(rules) > 0 {
			for _, rule := range rules {
				segments := strings.Split(rule, ";")
				if len(segments) > 0 {
					parsedMetrics = append(parsedMetrics, segments[0])
				}
			}
			exp.parsedMetrics = parsedMetrics
		}
	}
	return exp.parsedMetrics
}

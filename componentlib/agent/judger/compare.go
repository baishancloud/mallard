package judger

import (
	"math"
	"strings"
)

// CompareFunc is function to run comparation of left and right value
type CompareFunc func(left, right float64) bool

// NewCompareFunc return compare functions with operator name
func NewCompareFunc(operator string) CompareFunc {
	switch operator {
	case "=", "==":
		return func(left, right float64) bool {
			return math.Abs(left-right) < 1e-5
		}
	case "!=":
		return func(left, right float64) bool {
			return math.Abs(left-right) > 1e-5
		}
	case "<":
		return func(left, right float64) bool {
			return left < right
		}
	case "<=":
		return func(left, right float64) bool {
			return left <= right
		}
	case ">":
		return func(left, right float64) bool {
			return left > right
		}
	case ">=":
		return func(left, right float64) bool {
			return left >= right
		}
	}
	return nil
}

// TagFilterFunc is tag filter function
type TagFilterFunc func(left string) bool

// NewTagFilterFunc parse operator to gain right tag string and proper filter func
func NewTagFilterFunc(key, operator string) (string, TagFilterFunc) {
	opList := parseOperatorToList(operator)
	if len(opList) == 0 {
		return singleValueFilter(key, operator)
	}
	if strings.HasSuffix(key, "^") {
		return strings.TrimSuffix(key, "^"), func(left string) bool {
			for _, op := range opList {
				if strings.HasPrefix(left, op) {
					return true
				}
			}
			return false
		}
	}
	if strings.HasSuffix(key, "$") {
		return strings.TrimSuffix(key, "$"), func(left string) bool {
			for _, op := range opList {
				if strings.HasSuffix(left, op) {
					return true
				}
			}
			return false
		}
	}
	if strings.HasSuffix(key, "*") {
		return strings.TrimSuffix(key, "*"), func(left string) bool {
			for _, op := range opList {
				if strings.Contains(left, op) {
					return true
				}
			}
			return false
		}
	}
	if strings.HasSuffix(key, "!") {
		return strings.TrimSuffix(key, "!"), func(left string) bool {
			isMatch := 0
			for _, op := range opList {
				if left != op {
					isMatch++
				}
			}
			return isMatch == len(opList)
		}
	}
	return strings.TrimSuffix(key, "="), func(left string) bool {
		for _, op := range opList {
			if left == op {
				return true
			}
		}
		return false
	}
}

func singleValueFilter(key string, operator string) (string, TagFilterFunc) {
	if strings.HasSuffix(key, "^") {
		return strings.TrimSuffix(key, "^"), func(left string) bool {
			return strings.HasPrefix(left, operator)
		}
	}
	if strings.HasSuffix(key, "$") {
		return strings.TrimSuffix(key, "$"), func(left string) bool {
			return strings.HasSuffix(left, operator)
		}
	}
	if strings.HasSuffix(key, "*") {
		return strings.TrimSuffix(key, "*"), func(left string) bool {
			return strings.Contains(left, operator)
		}
	}
	if strings.HasSuffix(key, "!") {
		return strings.TrimSuffix(key, "!"), func(left string) bool {
			return left != operator && left != ""
		}
	}
	return strings.TrimSuffix(key, "="), func(left string) bool {
		return left == operator
	}
}

func parseOperatorToList(op string) []string {
	if !strings.HasPrefix(op, "[") {
		return nil
	}
	op = strings.TrimPrefix(op, "[")
	op = strings.TrimSuffix(op, "]")
	op = strings.TrimSpace(op)
	op = strings.Trim(op, "|")
	return strings.Split(op, "|")
}

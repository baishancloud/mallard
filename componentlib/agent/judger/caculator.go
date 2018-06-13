package judger

import (
	"errors"
	"math"

	"github.com/baishancloud/mallard/corelib/models"
)

func getCalculateLimit(l int, t string) int {
	switch t {
	case "diff":
		return l + 1
	case "diffavg":
		return l + 1
	case "adiff":
		return l + 1
	case "adiffavg":
		return l + 1
	case "rdiff":
		return l + 1
	case "rdiffavg":
		return l + 1
	case "pdiff":
		return l + 1
	case "pdiffavg":
		return l + 1
	}
	return l
}

// CalculateFunc is calculate function to compare each values in history data with right value
type CalculateFunc func(
	values []*models.EventValue,
	rightValue float64,
	compareFn CompareFunc,
	args ...interface{}) (float64, bool, error)

// NewCalculateFunc return calculate function with type name
func NewCalculateFunc(t string) CalculateFunc {
	switch t {
	case "all":
		return calculateAll
	case "diff":
		return calculateDiff
	case "diffavg":
		return calculateDiffAvg
	case "adiff":
		return calculateAdiff
	case "adiffavg":
		return calculateAdiffAvg
	case "rdiff":
		return calculateRdiff
	case "rdiffavg":
		return calculateRdiffAvg
	case "pdiff":
		return calculatePdiff
	case "pdiffavg":
		return calculatePdiffAvg
	case "avg":
		return calculateAvg
	case "sum":
		return calculateSum
	case "have":
		return calculateHave
	}
	return nil
}

func calculateAll(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	isOK := true
	for _, v := range values {
		if !compareFn(v.Value, rightValue) {
			return v.Value, false, nil
		}
	}
	return values[0].Value, isOK, nil
}

func calculateDiff(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValues []float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		realValues = append(realValues, values[0].Value-v.Value)
	}
	var (
		isOK    int
		leftIdx = -1
	)
	for i, v := range realValues {
		if compareFn(v, rightValue) {
			isOK++
			if leftIdx < 0 {
				leftIdx = i
			}
			//leftValue = v
		}
	}
	if leftIdx < 0 {
		leftIdx = 0
	}
	return realValues[leftIdx], isOK == len(realValues), nil
}

func calculateDiffAvg(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValue float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		realValue += values[0].Value - v.Value
	}
	realValue = realValue / float64(len(values)-1)
	return realValue, compareFn(realValue, rightValue), nil
}

func calculateRdiff(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValues []float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		realValues = append(realValues, values[i-1].Value-v.Value)
	}
	var (
		isOK    int
		leftIdx = -1
	)
	for i, v := range realValues {
		if compareFn(v, rightValue) {
			isOK++
			if leftIdx < 0 {
				leftIdx = i
			}
			//leftValue = v
		}
	}
	if leftIdx < 0 {
		leftIdx = 0
	}
	return realValues[leftIdx], isOK == len(realValues), nil
}

func calculateRdiffAvg(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValue float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		realValue += (values[i-1].Value - v.Value)
	}
	realValue = realValue / float64(len(values)-1)
	return realValue, compareFn(realValue, rightValue), nil
}

func calculateAdiff(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValues []float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		realValues = append(realValues, math.Abs(v.Value-values[0].Value))
	}
	var (
		isOK    int
		leftIdx = -1
	)
	for i, v := range realValues {
		if compareFn(v, rightValue) {
			isOK++
			// leftValue = v
			if leftIdx < 0 {
				leftIdx = i
			}
		}
	}
	if leftIdx < 0 {
		leftIdx = 0
	}
	return realValues[leftIdx], isOK == len(realValues), nil
}

func calculateAdiffAvg(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValue float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		realValue += math.Abs(values[0].Value - v.Value)
	}
	realValue = realValue / float64(len(values)-1)
	return realValue, compareFn(realValue, rightValue), nil
}

func minValue(v1, v2 float64) float64 {
	if v1 > v2 {
		return v2
	}
	return v1
}

func calculatePdiff(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValues []float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		value := values[i-1].Value - v.Value
		value = value / minValue(values[i-1].Value, v.Value) * 100
		realValues = append(realValues, value)
	}
	var (
		isOK    int
		leftIdx = -1
	)
	for i, v := range realValues {
		if compareFn(v, rightValue) {
			isOK++
			if leftIdx < 0 {
				leftIdx = i
			}
			//leftValue = v
		}
	}
	if leftIdx < 0 {
		leftIdx = 0
	}
	return realValues[leftIdx], isOK == len(realValues), nil
}

func calculatePdiffAvg(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValue float64
	for i, v := range values {
		if i == 0 {
			continue
		}
		value := values[i-1].Value - v.Value
		value = value / minValue(values[i-1].Value, v.Value) * 100
		realValue += value
	}
	realValue = realValue / float64(len(values)-1)
	return realValue, compareFn(realValue, rightValue), nil
}

func calculateSum(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValue float64
	for _, v := range values {
		realValue += v.Value
	}
	return realValue, compareFn(realValue, rightValue), nil
}

func calculateAvg(values []*models.EventValue, rightValue float64, compareFn CompareFunc, _ ...interface{}) (float64, bool, error) {
	var realValue float64
	for _, v := range values {
		realValue += v.Value
	}
	realValue = realValue / float64(len(values))
	return realValue, compareFn(realValue, rightValue), nil
}

var (
	// ErrHaveFuncNeedIntArgument means have function need one int argument
	ErrHaveFuncNeedIntArgument = errors.New("have func need 1 int argument")
)

func calculateHave(values []*models.EventValue, rightValue float64, compareFn CompareFunc, args ...interface{}) (float64, bool, error) {
	if len(args) < 1 {
		return 0, false, ErrHaveFuncNeedIntArgument
	}
	limit, ok := args[0].(int)
	if !ok {
		return 0, false, ErrHaveFuncNeedIntArgument
	}
	var (
		triggerCount      int
		leftValueIndex    = -1
		notLeftValueIndex = -1
	)
	for idx, v := range values {
		if compareFn(v.Value, rightValue) {
			triggerCount++
			if leftValueIndex < 0 {
				leftValueIndex = idx
			}
		} else {
			if notLeftValueIndex < 0 {
				notLeftValueIndex = idx
			}
		}
		if triggerCount >= limit {
			return values[leftValueIndex].Value, true, nil
		}
	}
	return values[notLeftValueIndex].Value, false, nil
}

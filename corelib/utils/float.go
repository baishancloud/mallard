package utils

import (
	"errors"
	"math"
	"strconv"
	"time"
)

// FixFloat returns fixed digitals precision float64 value
func FixFloat(f float64, digit ...int) float64 {
	var useDigit = 2
	if len(digit) > 0 {
		useDigit = digit[0]
	}
	pow := math.Pow10(useDigit)
	i := int64(f * pow)
	return float64(i) / pow
}

// ToFloat64 converts interface to float64 value
func ToFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, errors.New("nil value")
	}
	var (
		floatValue float64
		err        error
	)
	isValid := true
	switch value := v.(type) {
	case string:
		if floatValue, err = strconv.ParseFloat(value, 64); err != nil {
			isValid = false
		}
	case float64:
		floatValue = value
	case float32:
		floatValue = float64(value)
	case int:
		floatValue = float64(value)
	case int32:
		floatValue = float64(value)
	case int64:
		floatValue = float64(value)
	case uint:
		floatValue = float64(value)
	case uint32:
		floatValue = float64(value)
	case uint64:
		floatValue = float64(value)
	default:
		isValid = false
	}
	if !isValid {
		return 0, errors.New("wrong value")
	}
	return floatValue, nil
}

// DurationMS returns ms of duration
func DurationMS(du time.Duration) int64 {
	return du.Nanoseconds() / 1e6
}

// DurationSinceMS returns ms duration from t to now
func DurationSinceMS(t time.Time) int64 {
	return DurationMS(time.Since(t))
}

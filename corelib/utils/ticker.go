package utils

import "time"

// Ticker runs function in time loop,
// run func then wait next ticker
func Ticker(interval time.Duration, fn func()) {
	if fn == nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		fn()
		<-ticker.C
	}
}

// TickerThen runs function in time loop,
// wait next ticker then run func
func TickerThen(interval time.Duration, fn func()) {
	if fn == nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		fn()
	}
}

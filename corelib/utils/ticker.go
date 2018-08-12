package utils

import "time"

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

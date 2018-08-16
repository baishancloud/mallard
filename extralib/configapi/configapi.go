package configapi

import (
	"strings"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log = zaplog.Zap("configapi")

	centerAPI string
)

// SetAPI sets api url
func SetAPI(urlStr string, role ...string) {
	centerAPI = strings.TrimSuffix(urlStr, "/")
	log.Info("set-api", "url", centerAPI)
	if len(role) > 0 {
		httputil.SetSpecialRole(role[0])
	}
}

var (
	intervalsLock   sync.RWMutex
	intervals       = make(map[string]func())
	intervalFactory = make(map[string]func())
)

func registerFactory(name string, fn func()) {
	if fn == nil {
		return
	}
	intervalsLock.Lock()
	intervalFactory[name] = fn
	intervalsLock.Unlock()
}

// SetIntervals sets valid functions to run interval
func SetIntervals(names ...string) {
	intervalsLock.Lock()
	for _, name := range names {
		if fn := intervalFactory[name]; fn != nil {
			intervals[name] = fn
		}
	}
	intervalsLock.Unlock()
}

// Intervals starts to run intervals
func Intervals(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		intervalsLock.RLock()
		for _, fn := range intervals {
			if fn != nil {
				go fn()
			}
		}
		intervalsLock.RUnlock()
		<-ticker.C
	}
}

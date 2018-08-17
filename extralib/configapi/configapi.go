package configapi

import (
	"strings"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log = zaplog.Zap("configapi")

	centerAPI string
)

// IntervalOption is option for sync intervals
type IntervalOption struct {
	Types   []string
	Addr    string
	Role    string
	Service *models.HostService
}

// SetForInterval sets options for interval
func SetForInterval(opt IntervalOption) {
	setAPI(opt.Addr, opt.Role)
	setIntervals(opt.Types...)
	if opt.Service != nil {
		setHostService(opt.Service)
	}
}

func setAPI(urlStr string, role ...string) {
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

func setIntervals(names ...string) {
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
	fn := func() {
		intervalsLock.RLock()
		for _, fn := range intervals {
			if fn != nil {
				go fn()
			}
		}
		intervalsLock.RUnlock()
	}
	utils.Ticker(interval, fn)
}

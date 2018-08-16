package configapi

import (
	"fmt"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	strategiesLock     sync.RWMutex
	strategiesHash     string
	strategyNodataKeys []string
	strategies         = make(map[int]*models.Strategy)
	strategiesCounter  = expvar.NewBase("csdk.strategies")
)

const (
	TypeStrategies = "strategies"
)

func init() {
	registerFactory(TypeStrategies, reqStrategies)
	expvar.Register(strategiesCounter)
}

func reqStrategies() {
	url := centerAPI + "/api/strategy?gzip=1&hash=" + strategiesHash
	ss := make(map[int]*models.Strategy, 1e3)
	statusCode, hash, err := httputil.GetJSONWithHash(url, time.Second*10, &ss)
	if err != nil {
		log.Warn("req-strategies-error", "error", err)
		return
	}
	if statusCode == 304 {
		log.Info("req-strategies-304")
		return
	}
	strategiesLock.Lock()
	strategies = ss
	strategiesHash = hash
	var keys []string
	for _, s := range strategies {
		if s.NoData > 0 {
			key := fmt.Sprintf("s_%d", s.ID)
			keys = append(keys, key)
		}
	}
	strategyNodataKeys = keys
	log.Info("req-strategies-ok", "hash", hash, "len", len(strategies))
	strategiesCounter.Set(int64(len(strategies)))
	strategiesLock.Unlock()
}

// GetStrategyByID gets one strategy by id
func GetStrategyByID(id int) *models.Strategy {
	strategiesLock.RLock()
	defer strategiesLock.RUnlock()
	return strategies[id]
}

// GetStrategies gets all strategies
func GetStrategies() map[int]*models.Strategy {
	strategiesLock.RLock()
	defer strategiesLock.RUnlock()
	cp := make(map[int]*models.Strategy, len(strategies))
	for id, st := range strategies {
		cp[id] = st
	}
	return cp
}

// GetStrategyNodataKeys gets nodata keys
func GetStrategyNodataKeys() []string {
	return strategyNodataKeys
}

// GetStrategyNodata gets nodata strategies
func GetStrategyNodata() []*models.Strategy {
	strategiesLock.RLock()
	defer strategiesLock.RUnlock()
	var ss []*models.Strategy
	for _, s := range strategies {
		if s.NoData > 0 {
			ss = append(ss, s)
		}
	}
	return ss
}

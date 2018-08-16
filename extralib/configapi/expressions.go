package configapi

import (
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	expsLock    sync.RWMutex
	expsHash    string
	expsMap     = make(map[int]*models.Expression)
	expsCounter = expvar.NewBase("csdk.expressions")
)

const (
	TypeExpressions = "expressions"
)

func init() {
	registerFactory(TypeExpressions, reqExpressions)
	expvar.Register(expsCounter)
}

func reqExpressions() {
	url := centerAPI + "/api/expression?gzip=1&hash=" + expsHash
	ss := make(map[int]*models.Expression, 1e3)
	statusCode, hash, err := httputil.GetJSONWithHash(url, time.Second*10, &ss)
	if err != nil {
		log.Warn("req-exps-error", "error", err)
		return
	}
	if statusCode == 304 {
		log.Info("req-exps-304")
		return
	}
	expsLock.Lock()
	expsMap = ss
	expsHash = hash
	log.Info("req-exps-ok", "hash", hash, "len", len(expsMap))
	expsCounter.Set(int64(len(expsMap)))
	expsLock.Unlock()
}

// GetExpressionByID gets one expression by id
func GetExpressionByID(id int) *models.Expression {
	expsLock.RLock()
	defer expsLock.RUnlock()
	return expsMap[id]
}

// GetExpressions gets all strategies
func GetExpressions() map[int]*models.Expression {
	expsLock.RLock()
	defer expsLock.RUnlock()
	cp := make(map[int]*models.Expression, len(expsMap))
	for id, st := range expsMap {
		cp[id] = st
	}
	return cp
}

// CheckExpressionsCache checks hash to get latest expressions data
func CheckExpressionsCache(hash string) (map[int]*models.Expression, string) {
	if hash == expsHash {
		return nil, hash
	}
	return GetExpressions(), expsHash
}

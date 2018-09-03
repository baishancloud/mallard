package configapi

import (
	"errors"
	"html/template"
	"strings"
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
	expsTpls    = make(map[int]*StrategyTplCache)
	expsCounter = expvar.NewBase("csdk.expressions")
)

const (
	// TypeExpressions is request type of expressions
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
	triggerExpvar(statusCode, err)
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

	// reload templates cache
	tpls := make(map[int]*StrategyTplCache, len(ss))
	for id, st := range ss {
		tpl := &StrategyTplCache{
			Note: st.Note,
		}
		if !strings.Contains(st.Note, "{{") {
			tpl.NoneedRender = true
			tpls[id] = tpl
			continue
		}
		noteTmpl, err := template.New("note").Funcs(strategyTplFuncs).Parse(st.Note)
		if err != nil {
			log.Warn("expression-tpl-error", "error", err, "id", id, "note", st.Note)
		} else {
			tpl.Tpl = noteTmpl
		}
		tpls[id] = tpl
	}
	expsTpls = tpls

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

var (
	// ErrExpressionTemplateNil means template cache is nil
	ErrExpressionTemplateNil = errors.New("exp-tpl-nil")
)

// RenderExpressionTpl renders event with proper expression id
func RenderExpressionTpl(id int, event *models.EventFull) (string, error) {
	expsLock.RLock()
	defer expsLock.RUnlock()
	tplCache := expsTpls[id]
	if tplCache == nil {
		return "", ErrExpressionTemplateNil
	}
	return tplCache.Render(event)
}

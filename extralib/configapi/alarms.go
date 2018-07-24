package configapi

import (
	"strconv"
	"sync"
	"time"

	"github.com/baishancloud/mallard/componentlib/center/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	cacheAlarms        = new(sqldata.Alarms)
	cacheAlarmRequests = make(map[int]map[string]*sqldata.AlarmSendRequest)
	cacheAlarmsLock    sync.RWMutex

	messagesCount     = expvar.NewBase("csdk.alarm_messages")
	alarmsCount       = expvar.NewBase("csdk.alarm_alarms")
	alarmActionsCount = expvar.NewBase("csdk.alarm_actions")
)

func init() {
	registerFactory("alarms", reqAlarms)
	registerFactory("alarm-requests", reqAlarmRequests)
	expvar.Register(messagesCount, alarmsCount, alarmActionsCount)
}

func reqAlarms() {
	url := centerAPI + "/api/alarm/all?gzip=1&crc=" + strconv.FormatUint(uint64(cacheAlarms.CRC), 10)
	alarms := new(sqldata.Alarms)
	statusCode, err := httputil.GetJSON(url, time.Second*5, alarms)
	if err != nil {
		log.Warn("req-alarms-error", "error", err)
		return
	}
	if statusCode == 304 {
		log.Info("req-alarms-304")
		return
	}
	if alarms.CRC != cacheAlarms.CRC {
		cacheAlarms = alarms
		alarmsCount.Set(int64(len(alarms.ForStrategies)))
		alarmActionsCount.Set(int64(len(alarms.Actions)))
		log.Info("req-alarms-ok", "crc", cacheAlarms.CRC)
		return
	}
	log.Info("req-alarms-same")
}

// Alarms gets cached Alarms data
func Alarms() *sqldata.Alarms {
	return cacheAlarms
}

// AlarmActionForStrategy gets alarm action for one strategy
func AlarmActionForStrategy(sid int) *models.AlarmAction {
	if len(cacheAlarms.StrategyActions) > 0 {
		return cacheAlarms.StrategyActions[sid]
	}
	return nil
}

// AlarmTemplateForAction gets action template
func AlarmTemplateForAction(actionID int) *models.Template {
	if len(cacheAlarms.Templates) == 0 {
		return nil
	}
	return cacheAlarms.Templates[actionID]
}

func reqAlarmRequests() {
	url := centerAPI + "/api/alarm/requests?gzip=1&crc=" + strconv.FormatUint(uint64(cacheAlarms.CRC), 10)
	requests := make(map[int]map[string]*sqldata.AlarmSendRequest)
	statusCode, err := httputil.GetJSON(url, time.Second*5, &requests)
	if err != nil {
		log.Warn("req-alarm-requests", "error", err)
		return
	}
	if statusCode == 304 {
		log.Info("req-alarm-requests-304")
		return
	}
	if len(requests) == 0 {
		log.Warn("req-alarm-requests-0")
		return
	}
	cacheAlarmsLock.Lock()
	var count int
	for _, reqs := range requests {
		count += len(reqs)
	}
	messagesCount.Set(int64(count))
	cacheAlarmRequests = requests
	log.Info("req-alarm-requests", "count", len(requests), "len", count)
	cacheAlarmsLock.Unlock()
}

// AlarmRequestByStrategy gets one alarm requests by strategy id
func AlarmRequestByStrategy(sid int) map[string]*sqldata.AlarmSendRequest {
	cacheAlarmsLock.RLock()
	defer cacheAlarmsLock.RUnlock()
	return cacheAlarmRequests[sid]
}

// AlarmForStrategy gets one alarm data for one strategy,
// contains uic, template and alarm users status
func AlarmForStrategy(sid int) *sqldata.AlarmsForStrategy {
	cacheAlarmsLock.RLock()
	defer cacheAlarmsLock.RUnlock()
	if cacheAlarms != nil {
		return cacheAlarms.ForStrategies[sid]
	}
	return nil
}

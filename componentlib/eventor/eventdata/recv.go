package eventdata

import (
	"errors"
	"strings"
	"time"

	"github.com/baishancloud/mallard/componentlib/eventor/redisdata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/configapi"
)

var (
	log = zaplog.Zap("eventdata")

	alarmCount           = expvar.NewDiff("rd.alarm_all")
	alarmOKCount         = expvar.NewDiff("rd.alarm_ok")
	alarmNOCount         = expvar.NewDiff("rd.alarm_no")
	alarmClosedCount     = expvar.NewDiff("rd.alarm_closed")
	alarmOutdatedCount   = expvar.NewDiff("rd.alarm_outdated")
	alarmNODATACount     = expvar.NewDiff("rd.alarm_nodata")
	alarmHappenCount     = expvar.NewBase("rd.alarm_happening")
	alarmHandleFailCount = expvar.NewDiff("rd.alarm_handle_fail")

	recvOKCount      = expvar.NewDiff("rd.recv_ok")
	recvProblemCount = expvar.NewDiff("rd.recv_problem")
	recvClosed       = expvar.NewDiff("rd.recv_closed")
	recvOutdated     = expvar.NewDiff("rd.recv_outdated")
	recvMaintains    = expvar.NewDiff("rd.recv_maintains")
	recvOuttime      = expvar.NewDiff("rd.recv_outtime")
	recvExpired      = expvar.NewDiff("rd.recv_expired")
)

func init() {
	expvar.Register(alarmCount, alarmOKCount, alarmNOCount,
		alarmClosedCount, alarmOutdatedCount,
		alarmNODATACount, alarmHappenCount,
		alarmHandleFailCount,
		recvOKCount, recvProblemCount,
		recvClosed, recvMaintains, recvOutdated, recvOuttime, recvExpired)
}

// Receive receives events array and handle
func Receive(events []*models.Event) {
	var (
		err    error
		length = len(events)
	)
	if length == 0 {
		log.Warn("recv-0")
		return
	}

	// check events to alarm
	t := time.Now().Unix()
	var (
		okCount    int64
		closeCount int64
	)
	for _, event := range events {
		if event == nil {
			continue
		}
		var (
			operateCode int
		)
		switch event.Status {
		case models.EventOutdated:
			recvOutdated.Incr(1)
			operateCode = opAlarm
			Remove(event.ID)
			log.Debug("recv-outdated", "eid", event.ID)
		case models.EventClosed:
			closeCount++
			operateCode = opIgore
			if isInMemory(event.ID) {
				operateCode = opAlarm
				Remove(event.ID)
				log.Info("recv-closed-alarm", "eid", event.ID)
			}
		default:
			if event.Status == models.EventOk {
				okCount++
			}
			operateCode, err = Check(event)
			if err != nil {
				log.Warn("check-event-error", "error", err, "status", event.Status.String(), "event", event)
			}
		}
		if operateCode != 0 {
			if operateCode == opExired {
				recvExpired.Incr(1)
				continue
			}
			if err = Handle(event, operateCode, t); err != nil {
				alarmHandleFailCount.Incr(1)
				log.Warn("handle-event-error", "error", err, "status", event.Status.String(), "event", event)
			}
		}
	}
	recvOKCount.Incr(okCount)
	recvClosed.Incr(closeCount)

	// save events to redis
	_, err = redisdata.CacheEvents(events)
	if err != nil {
		log.Warn("cache-events-error", "error", err, "len", length)
	}

	// update nodata records
	if err = cacheNODATA(events); err != nil {
		log.Warn("cache-nodata-error", "error", err, "len", length)
	}
}

const (
	opIgore  = 1
	opAlarm  = 3
	opUpdate = 5
	opExired = 7
)

var (
	// ErrStrategyZeroNodataKeys means zero nodata
	ErrStrategyZeroNodataKeys = errors.New("zero-nodata-keys")
)

func cacheNODATA(events []*models.Event) error {
	nodataKeys := configapi.GetStrategyNodataKeys()
	if len(nodataKeys) == 0 {
		return ErrStrategyZeroNodataKeys
	}
	setNodataMap := make(map[string]int64)
	removeNodataMap := make(map[string]int64)
	for _, evt := range events {
		// only for s_eid events
		if !strings.HasPrefix(evt.ID, "s_") {
			continue
		}
		idx := strings.LastIndex(evt.ID, "_")
		if idx < 2 {
			continue
		}
		// special status, clean nodata
		if evt.Status == models.EventOutdated || evt.Status == models.EventClosed {
			removeNodataMap[evt.ID] = 1
			continue
		}
		prefix := evt.ID[:idx]
		for _, s := range nodataKeys {
			if prefix == s {
				log.Debug("nodata-set", "eid", evt.ID)
				setNodataMap[evt.ID] = evt.Time

			}
		}
	}
	if len(setNodataMap) > 0 {
		if err := redisdata.SetSomeEventNoData(setNodataMap); err != nil {
			log.Warn("nodata-set-error", "count", len(setNodataMap), "error", err)
		} else {
			log.Info("nodata-set-ok", "count", len(setNodataMap), "all", len(events))
		}
	}
	if len(removeNodataMap) > 0 {
		if err := redisdata.RemoveSomeEventNodata(removeNodataMap); err != nil {
			log.Warn("nodata-set-rm-error", "count", len(removeNodataMap), "error", err)
		} else {
			log.Info("nodata-set-rm-ok", "count", len(removeNodataMap), "all", len(events))
		}
	}
	return nil
}

// Check checks event to determine operation code
func Check(event *models.Event) (int, error) {
	isInProblem := isInMemory(event.ID)
	if event.Status == models.EventOk {
		// if event is ok and in problem, reset alarm time
		if isInProblem {
			// remove problem message and set to alarm to cleanup event
			if err := Remove(event.ID); err != nil {
				log.Warn("check-remove-event-error", "eid", event.ID, "status", event.Status.String(), "error", err)
			}
			return opAlarm, nil
		}
		return opIgore, nil
	}
	// if event is problem,...
	if event.Status == models.EventProblem {
		recvProblemCount.Incr(1)
		// if endpoint is in maintain,
		// ignore any problem event from the endpoint
		if redisdata.CheckEndpointMaintain(event.Endpoint) {
			log.Info("ignore-maintain", "status", event.Status.String(), "eid", event.ID, "endpoint", event.Endpoint)
			recvMaintains.Incr(1)
			return opIgore, nil
		}
		if t2 := redisdata.CheckRawEventTime(event.ID, event.Time); t2 > 0 {
			log.Info("event-expired", "status", event.Status.String(), "eid", event.ID, "endpoint", event.Endpoint, "t0", event.Time, "t2", t2)
			recvExpired.Incr(1)
			return opExired, nil
		}
		// if event is in problem , set opCode to 'update'
		if isInProblem {
			// update problem message in redis
			if err := SaveAlarming(event, false); err != nil {
				log.Warn("save-event-error", "eid", event.ID, "status", event.Status.String(), "error", err)
			}
			return opUpdate, nil
		}
		if err := SaveAlarming(event, true); err != nil {
			log.Warn("save-event-error", "eid", event.ID, "status", event.Status.String(), "error", err)
		}
		return opAlarm, nil
	}
	return opIgore, nil
}

// Handle handles event with proper operate code
func Handle(event *models.Event, op int, t int64) error {
	if op == opIgore || op == opExired {
		return nil
	}
	fullEvent, note, err := Convert(event, true)
	if err != nil {
		if fullEvent == nil {
			return err
		}
		log.Warn("event-handle-error", "status", event.Status.String(), "eid", event.ID, "error", err)
	}
	if err = redisdata.SetAlarmingNote(event.ID, note); err != nil {
		return err
	}
	if op == opUpdate {
		// update event notes for web portal
		if err := redisdata.RefreshAlarmingEvent(fullEvent, note); err != nil {
			log.Warn("event-update-note-error", "status", event.Status.String(), "eid", event.ID, "error", err)
		}
		log.Debug("event-update", "status", event.Status.String(), "eid", event.ID, "note", note, "ep", event.Endpoint)
		return nil
	}

	if st, ok := fullEvent.Strategy.(*models.Strategy); ok && event.Status == models.EventProblem {
		// if strategy is not in time range, stop resolving the event
		if !st.IsInTime(t) {
			Remove(event.ID)
			log.Info("event-out-time", "status", event.Status.String(), "eid", event.ID, "ep", event.Endpoint)
			recvOuttime.Incr(1)
			return nil
		}
	}

	log.Info("event-alarm", "status", event.Status.String(), "eid", event.ID, "event", fullEvent, "ep", event.Endpoint)
	// send event to alarm queue
	Alarm(fullEvent)
	return nil
}

// Alarm sends events to alarming queue
func Alarm(fullEvent *models.EventFull) {
	alarmCount.Incr(1)
	switch fullEvent.Status {
	case models.EventOk.String():
		alarmOKCount.Incr(1)
	case models.EventProblem.String():
		alarmNOCount.Incr(1)
	case models.EventClosed.String():
		alarmClosedCount.Incr(1)
	case models.EventOutdated.String():
		alarmOutdatedCount.Incr(1)
	}
	if strings.HasPrefix("nodata_", fullEvent.ID) {
		alarmNODATACount.Incr(1)
	}

	queue, vlen, err := redisdata.PushAlarmEvent(fullEvent)
	if err != nil {
		log.Warn("lpush-error", "eid", fullEvent.ID, "status", fullEvent.Status, "ep", fullEvent.Endpoint)
		return
	}
	log.Info("lpush-ok", "eid", fullEvent.ID, "queue", queue, "length", vlen)
}

// Remove removes one event from memory and redis
func Remove(eid string) error {
	if err := redisdata.RemoveAlarmingEvent(eid); err != nil {
		return err
	}
	removeMemory(eid)
	return nil
}

// SaveAlarming saves event to memory and redis,
// if firstRecord, updates last-data
func SaveAlarming(event *models.Event, firstRecord bool) error {
	if err := redisdata.SaveAlarmingEvent(event, firstRecord); err != nil {
		return err
	}
	saveMemory(event)
	return nil
}

package eventdata

import (
	"errors"
	"strings"
	"time"

	"github.com/baishancloud/mallard/componentlib/compute/redisdata"
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

	recvOKCount   = expvar.NewDiff("rd.recv_ok")
	recvNOCount   = expvar.NewDiff("rd.recv_no")
	recvClosed    = expvar.NewDiff("rd.recv_closed")
	recvOutdated  = expvar.NewDiff("rd.recv_outdated")
	recvMaintains = expvar.NewDiff("rd.recv_maintains")
	recvOuttime   = expvar.NewDiff("rd.recv_outtime")
)

func init() {
	expvar.Register(alarmCount, alarmOKCount, alarmNOCount,
		alarmClosedCount, alarmOutdatedCount,
		alarmNODATACount, alarmHappenCount,
		alarmHandleFailCount,
		recvOKCount, recvNOCount,
		recvClosed, recvMaintains, recvOutdated, recvOuttime)
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
	if err = redisdata.CacheEvents(events); err != nil {
		log.Warn("cache-events-error", "error", err, "len", length)
	}
	if err = cacheNODATA(events); err != nil {
		log.Warn("cache-nodata-error", "error", err, "len", length)
	}

	t := time.Now().Unix()
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
			recvClosed.Incr(1)
			operateCode = opIgore
			if isInMemory(event.ID) {
				operateCode = opAlarm
				if err := Remove(event.ID); err != nil {
					log.Warn("recv-closed-remove-error", "error", err, "eid", event.ID)
				}
				log.Info("recv-closed-alarm", "eid", event.ID)
			} else {
				log.Debug("recv-closed", "eid", event.ID)
			}
		default:
			operateCode, err = Check(event)
			if err != nil {
				log.Warn("check-event-error", "error", err, "status", event.Status.String(), "event", event)
			}
		}
		if operateCode != 0 {
			if err = Handle(event, operateCode, t); err != nil {
				alarmHandleFailCount.Incr(1)
				log.Warn("handle-event-error", "error", err, "status", event.Status.String(), "event", event)
			}
		}
	}
}

const (
	opIgore  = 1
	opAlarm  = 3
	opUpdate = 5
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
	var count int
	for _, evt := range events {
		// only for s_eid events
		if !strings.HasPrefix(evt.ID, "s_") {
			continue
		}
		idx := strings.LastIndex(evt.ID, "_")
		if idx < 2 {
			continue
		}
		prefix := evt.ID[:idx]
		for _, s := range nodataKeys {
			if prefix == s {
				log.Debug("nodata-set", "eid", evt.ID)
				if err := redisdata.SetEventNodata(evt.ID, evt.Time); err != nil {
					log.Warn("nodata-set-error", "eid", evt.ID, "error", err)
				}

				count++
			}
		}
	}
	if count > 0 {
		log.Info("nodata-set-ok", "count", count, "all", len(events))
	}
	return nil
}

// Check checks event to determine operation code
func Check(event *models.Event) (int, error) {
	isInProblem := isInMemory(event.ID)
	if event.Status == models.EventOk {
		recvOKCount.Incr(1)
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
		recvNOCount.Incr(1)
		// if endpoint is in maintain,
		// ignore any problem event from the endpoint
		if redisdata.CheckEndpointMaintain(event.Endpoint) {
			log.Debug("ignore-maintain", "status", event.Status.String(), "eid", event.ID)
			recvMaintains.Incr(1)
			return opIgore, nil
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
	if op == opIgore {
		return nil
	}
	fullEvent, note, err := Convert(event, true)
	if err != nil {
		return err
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

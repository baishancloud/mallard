package judger

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	units       = make(map[int]*StrategyUnit)
	unitsAccept = make(map[string][]int)
	unitsLock   sync.RWMutex
	log         = zaplog.Zap("judger")

	eventCurrent = NewCurrent()

	strategyCount     = expvar.NewBase("event.strategy")
	eventOKCount      = expvar.NewDiff("events.ok")
	eventProblemCount = expvar.NewDiff("events.problem")
)

func init() {
	expvar.Register(strategyCount, eventOKCount, eventProblemCount)
}

// SetStrategyData sets strategies data
// if some unit closed, return closed events
func SetStrategyData(ss []*models.Strategy) []*models.Event {
	strategies := make(map[int]*models.Strategy, len(ss))
	for _, s := range ss {
		strategies[s.ID] = s
	}

	unitsLock.Lock()
	defer unitsLock.Unlock()

	var closedIDs []int

	// closed no-using unit
	before := len(units)
	for key := range units {
		if s := strategies[key]; s == nil { // not found in new stragies map, delete old
			unit := units[key]
			if unit != nil {
				delete(units, key)
				closedIDs = append(closedIDs, unit.ID())
				log.Debug("unit-closed", "id", unit.ID())
			}
		}
	}

	// update all units
	accepts := make(map[string][]int)
	for key, s := range strategies {
		u, ok := units[key]
		if ok {
			if err := u.SetStrategy(s); err != nil {
				log.Warn("unit-reload-error", "error", err, "id", u.ID())
			} else {
				accepts[s.Metric] = append(accepts[s.Metric], s.ID)
			}
		} else {
			var err error
			u, err = NewUnit(s)
			if err == nil {
				units[key] = u
				accepts[s.Metric] = append(accepts[s.Metric], s.ID)
				// log.Debug("unit-new", "id", u.ID())
			} else {
				log.Warn("unit-new-error", "error", err, "s", s)
			}
		}
	}
	for k := range accepts {
		sort.Sort(sort.IntSlice(accepts[k]))
	}
	unitsAccept = accepts

	log.Debug("set-strategy", "before", before, "after", len(units))
	strategyCount.Set(int64(len(units)))

	return genClosedEvents(closedIDs)
}

func genClosedEvents(closed []int) []*models.Event {
	var closedEvents []*models.Event
	var now = time.Now().Unix()
	for _, sid := range closed {
		events := eventCurrent.FindByStrategy(sid)
		for _, evt := range events {
			evt.Status = models.EventClosed
			evt.Time = now
			evt.Step = 0
			closedEvents = append(closedEvents, evt)
			log.Info("event-close", "eid", evt.ID)
		}
	}
	return closedEvents
}

// Judge check metrics to generate events
func Judge(metrics []*models.Metric) []*models.Event {
	events := make([]*models.Event, 0, len(metrics))
	for _, metric := range metrics {
		evts := judgeOnce(metric)
		if len(evts) == 0 {
			continue
		}
		events = append(events, evts...)
	}
	if len(events) > 0 {
		log.Info("judge-ok", "metrics", len(metrics), "events", len(events))
	}
	return events
}

func judgeOnce(metric *models.Metric) []*models.Event {
	unitsLock.RLock()
	defer unitsLock.RUnlock()

	unitList := unitsAccept[metric.Name]
	if len(unitList) == 0 {
		return nil
	}

	now := time.Now().UnixNano() / 1e6
	events := make([]*models.Event, 0, len(unitList))
	for _, id := range unitList {
		u := units[id]
		if u == nil {
			continue
		}
		if !u.AcceptTag(metric.FullTags()) {
			continue
		}
		leftValue, status, err := u.Check(metric, "", false)
		if _, ok := err.(FieldMissingError); ok {
			log.Debug("strategy-field-miss", "sid", u.ID(), "msg", err)
			continue
		}
		if status == models.EventIgnore {
			if err != nil {
				log.Debug("strategy-check-fail", "sid", u.ID(), "error", err)
			}
			continue
		}
		/*if status == models.EventSyntax && err != nil {
			log.Debug("strategy-check-fail", "sid", u.ID(), "error", err)
			continue
		}*/

		event := &models.Event{
			ID:         fmt.Sprintf("s_%d_%s", u.ID(), metric.Hash()),
			Status:     status,
			Time:       metric.Time,
			Strategy:   u.ID(),
			LeftValue:  leftValue,
			History:    u.History(metric.Hash()),
			Tags:       metric.Tags,
			Cycle:      metric.Step,
			Fields:     metric.Fields,
			Endpoint:   metric.Endpoint,
			CreateTime: now,
		}
		if len(event.Fields) > 0 {
			if _, ok := event.Fields["value"]; !ok {
				event.Fields["value"] = metric.Value
			}
		}

		if eventCurrent.ShouldAlarm(event) {
			if event.Status == models.EventOk {
				eventOKCount.Incr(1)
				if event.Step > 3 {
					event = event.Simple()
				}
			}
			if event.Status == models.EventProblem {
				eventProblemCount.Incr(1)
				log.Info("problem-event", "event", event)
			}
			events = append(events, event)
			/*log.Info("event",
			"s", event.Status.String(),
			"eid", event.ID,
			"step", event.Step)*/
		}
	}
	return events
}

// Currents returns current events in manager
func Currents() *Current {
	return eventCurrent
}

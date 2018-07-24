package eventdata

import (
	"time"

	"github.com/baishancloud/mallard/componentlib/eventor/redisdata"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/configapi"
)

var (
	// OutatedAvailableExpire is expiry of outdated event is still available for the endpoint or agent_endpoint
	OutatedAvailableExpire int64 = 1800
	// OutdatedExpire is expiry of outdated event checker
	OutdatedExpire int64 = 3600 * 2
)

// ScanOutdated scans alarming events to outdated in time loop
func ScanOutdated(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		scanOutdatedOnce()
	}
}

func scanOutdatedOnce() {
	nowUnix := time.Now().Unix()
	eventTimes, err := redisdata.GetAlarmEventTimes()
	if err != nil {
		log.Warn("outdated-error", "error", err)
		return
	}
	if len(eventTimes) == 0 {
		log.Debug("outdated-zero")
		return
	}
	var events []*models.Event
	for eid, tUnix := range eventTimes {
		diff := nowUnix - tUnix
		log.Debug("scan-outdated", "eid", eid, "diff", diff)
		if diff >= OutatedAvailableExpire {
			evt := generateOutdatedAvaible(eid, tUnix)
			if evt != nil {
				events = append(events, evt)
				log.Info("scan-outdated-available-problem", "eid", eid, "diff", diff)
			}
		}
		if diff >= OutdatedExpire {
			evt := generateOutdated(eid, tUnix)
			if evt != nil {
				events = append(events, evt)
				log.Info("scan-outdated-problem", "eid", eid, "diff", diff)
			}
		}
	}
	if len(events) > 0 {
		Receive(events)
	}
	log.Info("scan-outdated-ok", "count", len(events), "all", len(eventTimes))
}

func generateOutdatedAvaible(eid string, tUnix int64) *models.Event {
	event, err := redisdata.GetAlarmingEvent(eid)
	if err != nil {
		log.Warn("outdated-get-event-error", "eid", eid, "error", err)
		return nil
	}
	endpoint := event.Endpoint
	if event.Tags["agent_endpoint"] != "" {
		log.Debug("outdated-change-endpoint", "eid", eid, "endpoint", endpoint, "agent_endpoint", event.Tags["agent_endpoint"])
		endpoint = event.Tags["agent_endpoint"]
	}
	if endpoint == "" {
		log.Warn("outdated-get-event-no-endpoint", "eid", eid)
		return nil
	}
	var isAvaible bool
	ep := configapi.EndpointConfig(endpoint)
	if ep != nil {
		if sid := getStrategyID(eid); sid > 0 {
			if ep.IsUsingStrategy(sid) {
				isAvaible = true
			}
		}
	}
	if !isAvaible {
		log.Info("outdate-disavailabe", "eid", eid, "endpoint", endpoint)
		event.Time = tUnix
		event.Status = models.EventOutdated
		return event
	}
	return nil
}

func generateOutdated(eid string, tUnix int64) *models.Event {
	event, err := redisdata.GetAlarmingEvent(eid)
	if err != nil {
		log.Warn("outdated-get-event-error", "eid", eid, "error", err)
		return nil
	}
	event.Time = tUnix
	event.Status = models.EventOutdated
	return event
}

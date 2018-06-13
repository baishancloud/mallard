package eventdata

import (
	"time"

	"github.com/baishancloud/mallard/componentlib/compute/redisdata"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	// OutdatedExpire is expiry of outdated event checker
	OutdatedExpire int64 = 3600 * 3
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

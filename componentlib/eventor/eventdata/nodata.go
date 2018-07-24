package eventdata

import (
	"strconv"
	"strings"
	"time"

	"github.com/baishancloud/mallard/componentlib/eventor/redisdata"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/configapi"
)

// ScanNodata scans nodata events in time loop
func ScanNodata(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		scanNodataOnce()
	}

}

func scanNodataOnce() {
	ss := configapi.GetStrategyNodata()
	if len(ss) == 0 {
		log.Warn("nodata-zero-ss")
		return
	}
	allNodataSet, err := redisdata.GetAllEventNodata()
	if err != nil {
		log.Warn("nodata-getall-error", "error", err)
		return
	}
	// generate duration map
	durations := make(map[string]int64, len(ss))
	for _, s := range ss {
		durations["s_"+strconv.Itoa(s.ID)] = int64(s.NoData)
	}

	events := make([]*models.Event, 0, len(allNodataSet))
	now := time.Now().Unix()

	for eid, tUnix := range allNodataSet {
		duration := getNodataDuration(eid, durations)
		rawEvent, err := redisdata.GetRawEvent(eid)
		if err != nil || rawEvent == nil {
			log.Warn("nodata-get-raw-error", "eid", eid, "error", err, "raw-nil", rawEvent == nil)
			duration = -1
		}
		if duration < 0 {
			if err = redisdata.RemoveEventNodata(eid); err != nil {
				log.Warn("nodata-del-error", "eid", eid, "error", err)
			}
			continue
		}
		diff := now - tUnix
		event := &models.Event{
			ID:       "nodata_" + eid,
			Time:     now,
			Status:   models.EventOk,
			Strategy: getStrategyID(eid),
			Endpoint: rawEvent.Endpoint,
		}
		log.Debug("scan-nodata", "eid", eid, "diff", diff)
		if diff >= OutdatedExpire {
			event.Status = models.EventOutdated
			log.Debug("nodata-outdated", "eid", event.ID, "diff", diff)
		} else {
			if diff >= duration {
				event.Status = models.EventProblem
				log.Debug("nodata-flag", "eid", eid, "diff", diff, "duration", duration)
			}
		}
		events = append(events, event)
	}
	if len(events) > 0 {
		Receive(events)
	}
	log.Info("scan-nodata-ok", "all", len(allNodataSet), "events", len(events))
}

func getNodataDuration(eid string, durations map[string]int64) int64 {
	for key, duration := range durations {
		if strings.HasPrefix(eid, key) {
			return duration
		}
	}
	return -1
}

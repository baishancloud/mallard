package judger

import (
	"fmt"
	"strings"
	"sync"

	"github.com/baishancloud/mallard/corelib/models"
)

// Current is event history records manager
// it checks event history to determine the event should alarm
type Current struct {
	records map[string]*models.Event
	lock    sync.RWMutex
}

// NewCurrent return new event history object
func NewCurrent() *Current {
	return &Current{
		records: make(map[string]*models.Event),
	}
}

// ShouldAlarm check event should be alarmed by history
// If first record of this event, alarm
// If event status changes, alarm (but not-enough status)
// If event problem in two mod, alarm
// If event ok in 10 mod, alarm
func (h *Current) ShouldAlarm(event *models.Event) bool {
	if event == nil {
		return false
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	lastEvent := h.records[event.ID]
	if lastEvent == nil {
		h.records[event.ID] = event
		return event.Status <= 1 // when status >= 1, it's internal event status code, not public status code
	}
	if lastEvent.Status != event.Status {
		h.records[event.ID] = event
		return event.Status <= 1
	}
	event.Step = lastEvent.Step + 1
	h.records[event.ID] = event
	return event.Status <= 1
}

// FindByStrategy find events in records with same strategy id
func (h *Current) FindByStrategy(id int) map[string]*models.Event {
	m := make(map[string]*models.Event)
	key := fmt.Sprintf("s_%d_", id)
	h.lock.RLock()
	for eid, evt := range h.records {
		if strings.HasPrefix(eid, key) {
			m[eid] = evt
		}
	}
	h.lock.RUnlock()
	return m
}

// All return all events in history
func (h *Current) All() map[string]*models.Event {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.records
}

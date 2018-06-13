package eventdata

import (
	"strconv"
	"strings"
	"sync"

	"github.com/baishancloud/mallard/componentlib/compute/redisdata"
	"github.com/baishancloud/mallard/corelib/models"
)

var (
	problemEvents = make(map[string]int)
	problemLock   sync.RWMutex
)

// InitMemory init memory value from redis
func InitMemory() {
	events, err := redisdata.GetAlarmsEvents()
	if err != nil {
		log.Warn("init-memory-error", "error", err)
		return
	}
	problemLock.Lock()
	problemEvents = events
	problemLock.Unlock()
}

func isInMemory(eid string) bool {
	problemLock.RLock()
	defer problemLock.RUnlock()
	return problemEvents[eid] > 0
}

func listMemoryByStratgy(sid int) []string {
	problemLock.RLock()
	defer problemLock.RUnlock()
	var eids []string
	prefix := "s_" + strconv.Itoa(sid) + "_"
	for eid := range problemEvents {
		if strings.HasPrefix(eid, prefix) {
			eids = append(eids, eid)
		}
	}
	return eids
}

func saveMemory(event *models.Event) {
	problemLock.Lock()
	problemEvents[event.ID] = 1
	alarmHappenCount.Set(int64(len(problemEvents)))
	problemLock.Unlock()
}

func removeMemory(eid string) {
	problemLock.Lock()
	delete(problemEvents, eid)
	problemLock.Unlock()
}

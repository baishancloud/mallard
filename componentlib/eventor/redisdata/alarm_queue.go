package redisdata

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	alarmLowQueue  []string
	alarmHighQueue []string
	popStopFlag    int64
)

// EventRecord is record that reading from redis alarm queue
type EventRecord struct {
	Event  *models.EventFull
	IsHigh bool
}

// SetAlarmQueues sets alarm Queues
func SetAlarmQueues(low, high []string) {
	alarmLowQueue = low
	alarmHighQueue = high
}

// Pop pops events from alarm queue
func Pop(ch chan<- EventRecord, interval time.Duration) {
	go popLoop(alarmHighQueue, true, ch, interval)
	go popLoop(alarmLowQueue, false, ch, interval)
}

func popLoop(queues []string, isHigh bool, ch chan<- EventRecord, interval time.Duration) {
	log.Info("pop", "queues", queues)
	for {
		if atomic.LoadInt64(&popStopFlag) > 0 {
			log.Info("pop-stop")
			return
		}
		event, err := popQueue(queues)
		if err != nil {
			log.Warn("pop-error", "queues", queues, "error", err)
			time.Sleep(interval)
			continue
		}
		if event == nil {
			time.Sleep(interval)
			continue
		}
		log.Info("pop", "eid", event.ID, "status", event.Status)
		if ch != nil {
			ch <- EventRecord{
				Event:  event,
				IsHigh: isHigh,
			}
		}
	}
}

func popQueue(queues []string) (*models.EventFull, error) {
	for _, queue := range queues {
		qLen := queueCli.LLen(queue).Val()
		if qLen == 0 {
			continue
		}
		data, err := queueCli.RPop(queue).Result()
		if err != nil {
			return nil, err
		}
		if len(data) < 2 {
			return nil, fmt.Errorf("invalid-data:%s", data)
		}
		event := new(models.EventFull)
		return event, json.Unmarshal([]byte(data), event)
	}
	return nil, nil
}

// StopPop stops pop operation
func StopPop() {
	atomic.StoreInt64(&popStopFlag, 1)
}

package redisdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	// ErrEventNoBytes means gets no bytes of event from redis
	ErrEventNoBytes = errors.New("hget-events-no-bytes")
)

// GetAlarmingEvent gets one alarming event
func GetAlarmingEvent(eid string) (*models.Event, error) {
	if queueCli == nil {
		return nil, ErrQueueNil
	}
	eBytes, err := queueCli.HGet("alarms", eid).Bytes()
	if err != nil {
		return nil, err
	}
	if len(eBytes) == 0 {
		return nil, ErrEventNoBytes
	}
	event := new(models.Event)
	if err := json.Unmarshal(eBytes, event); err != nil {
		return nil, err
	}
	return event, nil
}

// GetRawEvent gets raw event data
// if ok, event data may be simplified
func GetRawEvent(eid string) (*models.Event, error) {
	if cacheCli == nil {
		return nil, ErrCacheNil
	}
	eBytes, err := cacheCli.HGet(eid, "current").Bytes()
	if err != nil {
		return nil, err
	}
	if len(eBytes) == 0 {
		return nil, ErrEventNoBytes
	}
	event := new(models.Event)
	if err := json.Unmarshal(eBytes, event); err != nil {
		return nil, err
	}
	return event, nil
}

// CacheEvents saves events to cache redis db
func CacheEvents(events []*models.Event) error {
	if cacheCli == nil {
		return ErrCacheNil
	}
	pipe := cacheCli.Pipeline()
	defer pipe.Close()

	for _, event := range events {
		if event == nil {
			continue
		}
		b, err := json.Marshal(event)
		if err != nil {
			log.Warn("encode-error", "error", err, "eid", event.ID, "status", event.Status.String())
			continue
		}
		pipe.HSet(event.ID, "current", string(b))
		pipe.HSet(event.ID, "lastest_time", int64(event.Time))
	}
	_, err := pipe.Exec()
	return err
}

// CheckEndpointMaintain checkes endpoint is maintained
func CheckEndpointMaintain(endpoint string) bool {
	if queueCli == nil {
		return false
	}
	maintainTime, _ := queueCli.HGet("host_maintain", endpoint).Int64()
	if maintainTime > 0 {
		now := time.Now().Unix()
		if now < maintainTime {
			log.Info("get-endpoint-maintain", "endpoint", endpoint, "maintain", maintainTime)
			return true
		}
	}
	return false
}

// GetAlarmsEvents gets alarming events list
func GetAlarmsEvents() (map[string]int, error) {
	if queueCli == nil {
		return nil, ErrQueueNil
	}
	values, err := queueCli.HGetAll("alarms_time").Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]int, len(values))
	for key := range values {
		result[key] = 1
	}
	return result, nil
}

// GetAlarmEventTimes gets alarming events time
func GetAlarmEventTimes() (map[string]int64, error) {
	if queueCli == nil {
		return nil, ErrQueueNil
	}
	values, err := queueCli.HGetAll("alarms_time").Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(values))
	for key, tStr := range values {
		tUnix, _ := strconv.ParseInt(tStr, 10, 64)
		if tUnix == 0 {
			continue
		}
		result[key] = tUnix
	}
	return result, nil
}

var (
	alarmZsetLayout string

	// ErrAlarmZsetEmpty means alarms queue layout is blank
	ErrAlarmZsetEmpty = errors.New("alarm-zset-layout-empty")
)

// SetAlarmLayout sets alarm queue layout
func SetAlarmLayout(layout string) {
	alarmZsetLayout = layout
}

// PushAlarmEvent pushes alarm event to zset
func PushAlarmEvent(event *models.EventFull) (string, int64, error) {
	if queueCli == nil {
		return "", 0, ErrQueueNil
	}
	if alarmZsetLayout == "" {
		return "", 0, ErrAlarmZsetEmpty
	}
	b, _ := json.Marshal(event)
	queue := fmt.Sprintf(alarmZsetLayout, event.Priority())
	res := queueCli.LPush(queue, b)
	if err := res.Err(); err != nil {
		return "", 0, err
	}
	return queue, res.Val(), nil
}

// SetAlarmingNote sets alarm note
func SetAlarmingNote(eid, note string) error {
	if queueCli == nil {
		return ErrQueueNil
	}
	return queueCli.HSet("note", eid, note).Err()
}

// SaveAlarmingEvent saves alarm event data to display
func SaveAlarmingEvent(event *models.Event, firstRecord bool) error {
	if queueCli == nil {
		return ErrQueueNil
	}
	b, _ := json.Marshal(event)
	pipe := queueCli.Pipeline()
	defer pipe.Close()
	pipe.HSet("alarms", event.ID, string(b))
	pipe.HSet("alarms_time", event.ID, event.Time)
	if firstRecord {
		pipe.HSet("last_alarm_time", event.ID, time.Now().Unix())
	}
	_, err := pipe.Exec()
	return err
}

// RemoveAlarmingEvent removes alarm event
func RemoveAlarmingEvent(eid string) error {
	if queueCli == nil {
		return ErrQueueNil
	}
	pipe := queueCli.Pipeline()
	defer pipe.Close()
	pipe.HSet("last_alarm_time", eid, 0)
	pipe.HDel("alarms", eid)
	pipe.HDel("alarms_time", eid)
	_, err := pipe.Exec()
	return err
}

// RefreshAlarmingEvent refreshes event notes
func RefreshAlarmingEvent(fullEvent *models.EventFull, note string) error {
	if queueCli == nil {
		return ErrQueueNil
	}
	ikey := fmt.Sprintf("fitem_%s", fullEvent.ID)
	leftKey := fmt.Sprintf("leftvalue_%s", fullEvent.ID)
	nkey := fmt.Sprintf("note_%s", fullEvent.ID)
	leftValue, err := leftValueItem(fullEvent.LeftValue)
	if err != nil {
		return err
	}
	pipe := queueCli.Pipeline()
	defer pipe.Close()
	pipe.Set(nkey, note, time.Hour)
	pipe.Set(ikey, leftValue, time.Hour)
	pipe.Set(leftKey, leftValue, time.Hour)
	_, err = pipe.Exec()
	return err
}

// SetEventNodata sets nodata eid and updated time
func SetEventNodata(eid string, tUnix int64) error {
	if cacheCli == nil {
		return ErrCacheNil
	}
	return cacheCli.HSet("nodata", eid, tUnix).Err()
}

// RemoveEventNodata removes nodata eid and updated time
func RemoveEventNodata(eid string) error {
	if cacheCli == nil {
		return ErrCacheNil
	}
	return cacheCli.HDel("nodata", eid).Err()
}

// CacheDBSize get size of cache redis db
func CacheDBSize() (int64, error) {
	if cacheCli == nil {
		return 0, ErrCacheNil
	}
	res := cacheCli.DBSize()
	return res.Val(), res.Err()
}

// GetAllEventNodata gets all nodata time records
func GetAllEventNodata() (map[string]int64, error) {
	if cacheCli == nil {
		return nil, ErrCacheNil
	}
	values, err := cacheCli.HGetAll("nodata").Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(values))
	for key, tStr := range values {
		tUnix, _ := strconv.ParseInt(tStr, 10, 64)
		if tUnix == 0 {
			continue
		}
		result[key] = tUnix
	}
	return result, nil
}

func leftValueItem(item interface{}) (interface{}, error) {
	var rv interface{}
	rv = item
	switch item.(type) {
	case string:
	case []byte:
	case int:
		rv = etoFloat64String(fmt.Sprintf("%d", rv))
	case int64:
		rv = etoFloat64String(fmt.Sprintf("%f", rv))
	case float64:
		rv = etoFloat64String(fmt.Sprintf("%f", rv))
	case bool:
	case nil:
	default:
		bs, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		rv = etoFloat64String(string(bs))
	}
	return rv, nil
}

func etoFloat64String(value string) (rv string) {
	var rf float64
	_, err := fmt.Sscanf(value, "%e", &rf)
	if err != nil {
		rv = "0.0"
	} else {
		rv = fmt.Sprintf("%f", rf)
	}
	if strings.Contains(rv, ".") {
		rv = strings.TrimRight(rv, "0")
		rv = strings.TrimRight(rv, ".")
	}
	return rv
}

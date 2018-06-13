package redisdata

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/go-redis/redis"
)

// Alert alerts event
func Alert(record EventRecord) {
	event := record.Event
	if event.Status == models.EventOutdated.String() || event.Status == models.EventClosed.String() {
		RemoveEvent(event.ID)
		return
	}
	if isFromStrategy(event.ID) {
		HandleEvent(event, record.IsHigh)
	}
}

const (
	strategyEventPrefix  = "s_"
	strategyNodataPrefix = "nodata_s_"
)

func isFromStrategy(eid string) bool {
	return strings.HasPrefix(eid, strategyEventPrefix) || strings.HasPrefix(eid, strategyNodataPrefix)
}

// HandleEvent sets event info to redis
func HandleEvent(event *models.EventFull, isHigh bool) {
	if err := models.FillEventStrategy(event); err != nil {
		log.Warn("fill-strategy-error", "error", err, "event", event)
		return
	}
	st, _ := event.Strategy.(*models.Strategy)
	if st == nil {
		log.Warn("fill-strategy-nil", "event", event)
		return
	}

	redisOperateEvent(event, st)
}

// SaveEventFromStrategy saves event from strategy
func SaveEventFromStrategy(event *models.EventFull, st *models.Strategy) {
	dto := models.NewEventDto(event, st)
	lastStr, _ := queueCli.Get(dto.ID).Result()
	if len(lastStr) > 0 {
		lastTodo := new(models.EventDto)
		if err := json.Unmarshal([]byte(lastStr), lastTodo); err != nil {
			log.Warn("insert-error", "error", err, "eid", dto.ID)
			return
		}
		if lastTodo.Priority != dto.Priority {
			key := fmt.Sprintf("p%dtime", lastTodo.Priority)
			queueCli.ZRem(key, lastTodo.ID)
			queueCli.Del(lastTodo.ID)
			log.Debug("delete-change-priority",
				"last", lastTodo.Priority,
				"todo", dto.Priority,
				"eid", dto.ID)
		}
	}

	bs, err := json.Marshal(dto)
	if err != nil {
		log.Warn("zadd-encode-error", "error", err, "eid", dto.ID)
		return
	}
	key := fmt.Sprintf("p%dtime", dto.Priority)
	z := redis.Z{
		Score:  float64(dto.Timestamp),
		Member: dto.ID,
	}
	if err := queueCli.ZAdd(key, z).Err(); err != nil {
		log.Warn("zadd-error", "error", err, "dto", dto)
		redisFailCount.Incr(1)
	}
	if err := queueCli.Set(dto.ID, string(bs), 0).Err(); err != nil {
		log.Warn("set-error", "error", err, "dto", dto)
		redisFailCount.Incr(1)
	}
	log.Info("zadd-ok", "eid", dto.ID, "key", key)
}

func redisOperateEvent(event *models.EventFull, st *models.Strategy) {
	if event.Status == models.EventOk.String() {
		RemoveEvent(event.ID)
		return
	}
	SaveEventFromStrategy(event, st)
	ToSubscrible(event.ID)
}

// RemoveEvent clean one event in redis
func RemoveEvent(eid string) {
	// clean raw data
	rawData, _ := queueCli.Get(eid).Bytes()
	if len(rawData) == 0 {
		forceClean(eid)
	} else {
		dto := new(models.EventDto)
		if err := json.Unmarshal(rawData, dto); err != nil {
			log.Warn("decode-error", "eid", eid, "error", err)
			forceClean(eid)
		} else {
			key := fmt.Sprintf("p%dtime", dto.Priority)
			queueCli.ZRem(key, dto.ID)
			log.Debug("zrem", "id", dto.ID, "zrem", key)
		}
	}
	queueCli.Del(eid)
	log.Debug("del", "eid", eid)
}

func forceClean(eid string) {
	var isForceClean bool
	for _, queue := range alarmHighQueue {
		affect, _ := queueCli.ZRem(queue, eid).Result()
		if affect > 0 {
			log.Debug("rm-zset", "eid", eid, "zrem", queue)
			isForceClean = true
		}
	}
	if !isForceClean {
		for _, queue := range alarmLowQueue {
			affect, _ := queueCli.ZRem(queue, eid).Result()
			if affect > 0 {
				log.Debug("rm-zset", "eid", eid, "zrem", queue)
				isForceClean = true
			}
		}
	}
	if !isForceClean {
		log.Warn("rm-zset-fail", "eid", eid)
	}
}

// HasNoAlarmFlag checks event no-alarm flag
func HasNoAlarmFlag(eid string) bool {
	if queueCli == nil {
		return false
	}
	t, _ := queueCli.HGet("noalarm", eid).Int64()
	if t > 0 {
		log.Debug("get-NOALARM", "eid", eid, "t", t)
		queueCli.HDel("noalarm", eid)
		if time.Now().Unix()-t > 3600 {
			log.Warn("get-NOALARM-outdated", "eid", eid, "t", t)
			return false
		}
		return true
	}
	return false
}

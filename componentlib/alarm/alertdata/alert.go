package alertdata

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baishancloud/mallard/componentlib/eventor/redisdata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

const (
	// EventOutdatedFlag is flag value of outdated event
	EventOutdatedFlag = -1
	// EventClosedFlag is flag value if closed event
	EventClosedFlag = -1 // -3
)

var (
	alertEventQueue = make(chan redisdata.EventRecord, 1000)

	problemIDs  = make(map[string][]eventDbID)
	problemLock sync.RWMutex

	alertInsertCount     = expvar.NewDiff("alertdb.insert")
	alertInsertFailCount = expvar.NewDiff("alertdb.insert_fail")
	alertEndEventCount   = expvar.NewDiff("alertdb.end_event")
	alertRecvCount       = expvar.NewDiff("alertdb.recv")
	alertDbErrorCount    = expvar.NewDiff("alertdb.db_fail")
	alertProblemsCount   = expvar.NewBase("alertdb.problems")
)

func init() {
	expvar.Register(alertInsertCount, alertRecvCount,
		alertEndEventCount, alertDbErrorCount,
		alertProblemsCount, alertInsertFailCount)
}

type eventDbID struct {
	ID    int64  `json:"id,omitempty"`
	Table string `json:"table,omitempty"`
	Time  int64  `json:"time,omitempty"`
}

// Alert pushed event to saving channel
func Alert(record redisdata.EventRecord) {
	alertRecvCount.Incr(1)
	alertEventQueue <- record
}

// StreamAlert starts handling events saving channel
func StreamAlert() {
	for {
		record := <-alertEventQueue
		if record.Event == nil {
			continue
		}
		handle(record.Event)
	}
}

func handle(event *models.EventFull) {
	if event.Status == models.EventOutdated.String() {
		endEvent(event.ID, EventOutdatedFlag)
		return
	}
	if event.Status == models.EventClosed.String() {
		endEvent(event.ID, EventClosedFlag)
		return
	}

	InsertEvent(event)
}

var (
	insertEventSQL = "insert into %s(eid,field_transform,func,operator,right_value,note,max_step,priority,status,endpoint,left_value,current_step,event_time,judge,pushed_tags,expression_id,strategy_id,template_id,insert_time) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
)

var (
	// MaxNoteSize is max length of note string
	MaxNoteSize = 512
)

// InsertEvent inserts new event
func InsertEvent(event *models.EventFull) {
	alertInsertCount.Incr(1)
	if err := models.FillEventStrategy(event); err != nil {
		log.Warn("fill-strategy-error", "error", err, "event", event)
		alertInsertFailCount.Incr(1)
		return
	}
	st, ok := event.Strategy.(*models.Strategy)
	if !ok {
		log.Warn("fill-strategy-not-ok", "event", event)
		alertInsertFailCount.Incr(1)
		return
	}

	tags, _ := json.Marshal(event.PushedTags)
	tableName := EventTableName(event.EventTime)
	sql := fmt.Sprintf(insertEventSQL, tableName)
	stmt, err := tryPrepare(sql, time.Unix(event.EventTime, 0))
	if err != nil {
		log.Warn("prepare-error", "error", err, "event", event)
		alertInsertFailCount.Incr(1)
		return
	}
	defer stmt.Close()

	// fix note length
	note := st.Note
	if len(note) > MaxNoteSize {
		note = utils.Substr(note, MaxNoteSize)
		log.Warn("substr-note", "eid", event.ID, "note", st.Note, "to", note)
	}

	var stID, expID int
	if strings.HasPrefix(event.ID, "e_") {
		expID = st.ID
	} else {
		stID = st.ID
	}

	result, err := stmt.Exec(
		event.ID,
		st.FieldTransform,
		st.Func,
		st.Operator,
		st.RightValue,
		note,
		st.MaxStep,
		st.Priority,
		event.Status,
		event.Endpoint,
		event.LeftValue,
		event.CurrentStep,
		event.EventTime,
		event.Judge,
		string(tags),
		expID,
		stID,
		st.TemplateID,
		time.Now().Unix(),
	)
	if err != nil {
		log.Warn("insert-error", "error", err, "event", event)
		alertInsertFailCount.Incr(1)
		return
	}
	var insertID int64
	if result != nil {
		insertID, _ = result.LastInsertId()
		if event.Status == models.EventOk.String() {
			updateEventDuration(event, insertID, false)

			problemLock.Lock()
			delete(problemIDs, event.ID)
			problemLock.Unlock()

		} else {

			problemLock.Lock()
			problemIDs[event.ID] = append(problemIDs[event.ID], eventDbID{
				ID:    insertID,
				Table: tableName,
				Time:  event.EventTime, // time.Now().Unix(),
			})
			alertProblemsCount.Set(int64(len(problemIDs)))
			problemLock.Unlock()
		}
	}
	log.Debug("insert-ok", "status", event.Status, "eid", event.ID, "insert_id", insertID)
}

var (
	lastEventTimeSQL       = "select id,event_time from %s where id < %d and eid = '%s' and duration = 0 and status = 'PROBLEM' order by id desc limit 1"
	lastPrevEventTimeSQL   = "select id,event_time from %s where eid = '%s' and duration = 0 and status = 'PROBLEM' order by id desc limit 1"
	updateEventDurationSQL = "UPDATE %s SET duration = ? WHERE id = ?"
	updateEventLastIDsSQL  = "UPDATE %s SET problem_ids = ? WHERE id = ?"
)

func updateEventDuration(event *models.EventFull, eventID int64, isPrev bool) {
	tableName := EventTableName(event.EventTime)
	if isPrev {
		tableName = EventPrevTableName(event.EventTime)
	}
	query := fmt.Sprintf(lastEventTimeSQL, tableName, eventID, event.ID)
	result := clerkDB.QueryRow(query)
	var id int64
	var t int64
	if err := result.Scan(&id, &t); err != nil {
		if err.Error() == sql.ErrNoRows.Error() && !isPrev {
			updateEventDuration(event, eventID, true)
			return
		}
		log.Warn("read-old-error", "error", err, "event", event.ID, "is_prev", isPrev, "table", tableName)
		alertDbErrorCount.Incr(1)
		return
	}
	duration := event.EventTime - t
	if duration > 0 {
		log.Debug("update-duration", "table", tableName, "lastid", id, "id", eventID)
		query = fmt.Sprintf(updateEventDurationSQL, tableName)
		if _, err := clerkDB.Exec(query, duration, id); err != nil {
			log.Warn("update-duration-error", "error", err, "eid", id, "event", event.ID)
			alertDbErrorCount.Incr(1)
		}
		query = fmt.Sprintf(updateEventDurationSQL, EventTableName(event.EventTime))
		if _, err := clerkDB.Exec(query, duration, eventID); err != nil {
			log.Warn("update-duration-error", "error", err, "event", event.ID)
			alertDbErrorCount.Incr(1)
		}
		problemLock.RLock()
		problems := problemIDs[event.ID]
		problemLock.RUnlock()
		if len(problems) == 0 {
			return
		}
		pids := make([]string, 0, len(problems))
		for _, pro := range problems {
			pids = append(pids, strconv.FormatInt(pro.ID, 10))
		}
		pidsStr := strings.Join(pids, ",")
		query = fmt.Sprintf(updateEventLastIDsSQL, tableName)
		if _, err := clerkDB.Exec(query, pidsStr, eventID); err != nil {
			log.Warn("update-problem-ids-error", "error", err, "event", event.ID)
			alertDbErrorCount.Incr(1)
		} else {
			log.Debug("update-problem-ids", "id", eventID, "pids", pids)
		}
	}
}

var (
	updateOutdatedDurationSQL = "UPDATE %s SET duration = ? WHERE id = ?"
)

func endEvent(eid string, flag int) {
	alertEndEventCount.Incr(1)
	problemLock.RLock()
	pids := problemIDs[eid]
	delete(problemIDs, eid)
	problemLock.RUnlock()
	if len(pids) == 0 {
		return
	}
	for _, eIDObj := range pids {
		sql := fmt.Sprintf(updateOutdatedDurationSQL, eIDObj.Table)
		if _, err := clerkDB.Exec(sql, flag, eIDObj.ID); err != nil {
			log.Warn("update-outdated-error", "error", err, "eid", eid, "pid", eIDObj.ID, "table", eIDObj.Table)
			alertDbErrorCount.Incr(1)
			return
		}
		log.Debug("update-outdated", "eid", eid, "pid", eIDObj.ID, "table", eIDObj.Table, "flag", flag)
	}
}

var (
	// ErrDbNil means db is nil
	ErrDbNil = errors.New("db-nil")
)

func tryPrepare(sqlStr string, t time.Time) (*sql.Stmt, error) {
	if clerkDB == nil {
		return nil, ErrDbNil
	}
	stmt, err := clerkDB.Prepare(sqlStr)
	if err != nil {
		y, m, d := getWeekNumber(t)
		createTable(y, m, d)
		return clerkDB.Prepare(sqlStr)
	}
	return stmt, err
}

func createTable(y, m, d int) {
	originsql := "CREATE TABLE if not exists `event%04d%02d%02d` ( \n" +
		"`id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n" +
		"`eid` varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',\n" +
		"`field_transform` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`func` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`operator` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`right_value` varchar(64) COLLATE utf8_unicode_ci NOT NULL,\n" +
		"`note` varchar(2048) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`max_step` int(11) DEFAULT NULL,\n" +
		"`priority` int(11) DEFAULT NULL,\n" +
		"`status` varchar(16) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`endpoint` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`left_value` varchar(64) COLLATE utf8_unicode_ci NOT NULL,\n" +
		"`current_step` int(11) DEFAULT NULL,\n" +
		"`event_time` int(11) DEFAULT NULL,\n" +
		"`judge` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`pushed_tags` varchar(2048) COLLATE utf8_unicode_ci DEFAULT NULL,\n" +
		"`expression_id` int(11) DEFAULT NULL,\n" +
		"`strategy_id` int(11) DEFAULT NULL,\n" +
		"`template_id` int(11) DEFAULT NULL,\n" +
		"`duration` int(11) DEFAULT 0,\n" +
		"`insert_time` int(11) NOT NULL DEFAULT 0,\n" +
		"`problem_ids` varchar(128) COLLATE utf8_unicode_ci DEFAULT '',\n" +
		"PRIMARY KEY (`id`),\n" +
		"KEY `eid` (`eid`)\n" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;"

	sql := fmt.Sprintf(originsql, y, m, d)
	log.Info("create-table", "tablename", fmt.Sprintf("event%04d%02d%02d", y, m, d))
	_, err := clerkDB.Exec(sql)
	if err != nil {
		log.Warn("create-table-error", "error", err)
	}
}

func getWeekNumber(t time.Time) (int, int, int) {
	// t := time.Unix(event.EventTime, 0)
	dt := (int)(t.Weekday()) - 1
	if dt >= 0 {
		t = t.AddDate(0, 0, -dt)
	} else {
		t = t.AddDate(0, 0, -6)
	}
	y, m, d := t.Date()
	return y, int(m), d
}

// EventTableName generates events db
func EventTableName(t int64) string {
	tt := time.Unix(t, 0)
	y, m, d := getWeekNumber(tt)
	return fmt.Sprintf("event%04d%02d%02d", y, m, d)
}

// EventPrevTableName generates prev events db
func EventPrevTableName(t int64) string {
	tt := time.Unix(t-3600*24*7-3600, 0)
	y, m, d := getWeekNumber(tt)
	return fmt.Sprintf("event%04d%02d%02d", y, m, d)
}

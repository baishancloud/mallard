package alertdata

import (
	"fmt"
	"time"
)

// SimpleEventDB is simple value for db event
type SimpleEventDB struct {
	ID         int    `db:"id"`
	EID        string `db:"eid"`
	Status     string `db:"status"`
	StrategyID int    `db:"strategy_id"`
}

var (
	findSimpleEventsSQL = "select id,eid,status,strategy_id from %s where strategy_id = %d order by id asc"
)

// FindSimpleEvents finds simple events
func FindSimpleEvents(table string, strategyID int) ([]*SimpleEventDB, error) {
	sqlStr := fmt.Sprintf(findSimpleEventsSQL, table, strategyID)
	rows, err := clerkDB.Queryx(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*SimpleEventDB
	for rows.Next() {
		event := new(SimpleEventDB)
		if err = rows.StructScan(event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

var recentAlarmSQL = "SELECT id,eid,note,status,endpoint,event_time,insert_time,strategy_id FROM %s WHERE endpoint = ? AND strategy_id = ? AND event_time > ? ORDER BY event_time DESC"

// AlarmEventDB is db object of one alarm event
type AlarmEventDB struct {
	ID         int    `db:"id"`
	EID        string `db:"eid"`
	Note       string `db:"note"`
	Status     string `db:"status"`
	Endpoint   string `db:"endpoint"`
	EventTime  int64  `db:"event_time"`
	InsertTime int64  `db:"insert_time"`
	StrategyID int    `db:"strategy_id"`
}

// FindRecentEvents finds recents events with endpoint, strategy and time duration
func FindRecentEvents(endpoint string, strategyID int, timeLimit int64) ([]*AlarmEventDB, error) {
	t := time.Now().Unix() - timeLimit
	rows, err := clerkDB.Queryx(fmt.Sprintf(recentAlarmSQL, EventTableName(t)), endpoint, strategyID, t)
	if err != nil {
		return nil, err
	}
	var alarms []*AlarmEventDB
	for rows.Next() {
		al := new(AlarmEventDB)
		if err := rows.StructScan(al); err != nil {
			return nil, err
		}
		if al.Endpoint == endpoint && al.StrategyID == strategyID {
			alarms = append(alarms, al)
		}
	}
	return alarms, nil
}

package alertdata

import "fmt"

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

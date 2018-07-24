package alertdata

import (
	"encoding/json"
	"io/ioutil"

	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/jmoiron/sqlx"
)

var (
	clerkDB *sqlx.DB

	log = zaplog.Zap("alertdata")
)

// SetDB sets db object
func SetDB(clerk *sqlx.DB) {
	clerkDB = clerk
	log.Info("set-db", "driver", clerk.DriverName())
}

// DumpProblems dumps happenning problems ids
func DumpProblems(file string) {
	problemLock.Lock()
	b, _ := json.Marshal(problemIDs)
	ioutil.WriteFile(file, b, 0644)
	log.Info("dump-problems", "count", len(problemIDs))
	problemLock.Unlock()
}

// ReadProblems reads happenning problems ids
func ReadProblems(file string) {
	b, _ := ioutil.ReadFile(file)
	if len(b) == 0 {
		return
	}
	problemLock.Lock()
	if err := json.Unmarshal(b, &problemIDs); err != nil {
		log.Warn("read-problems-error", "error", err)
	} else {
		log.Info("read-problems", "count", len(problemIDs))
		alertProblemsCount.Set(int64(len(problemIDs)))
	}
	problemLock.Unlock()

}

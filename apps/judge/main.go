package main

import (
	"time"

	"github.com/baishancloud/mallard/componentlib/dataflow/judgestore"
	"github.com/baishancloud/mallard/componentlib/dataflow/judgestore/filter"
	"github.com/baishancloud/mallard/componentlib/httphandler/judgehandler"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log = zaplog.Zap("judge")
)

func main() {
	log.SetDebug(true)

	queue := make(chan []*models.Metric, 1e4)

	judgestore.SetDir("./datastore")
	go filter.SyncFile("filters.json", func(f filter.ForMetrics) {
		judgestore.SetFilters(f)
	}, time.Minute)
	judgestore.RunClean()
	go judgestore.ProcessMetrics(queue)

	judgehandler.SetQueue(queue)
	go httputil.Listen("0.0.0.0:21789", judgehandler.Create())

	go expvar.PrintAlways("mallard2_judge_perf", "", time.Minute)

	osutil.Wait()

	httputil.Close()
	judgestore.Close()
	log.Sync()
}

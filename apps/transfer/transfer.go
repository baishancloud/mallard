package main

import (
	"runtime"
	"time"

	"github.com/baishancloud/mallard/componentlib/dataflow/eventsender"
	"github.com/baishancloud/mallard/componentlib/dataflow/queues"
	"github.com/baishancloud/mallard/componentlib/httphandler/transferhandler"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/configapi"
)

var (
	version    = "2.5.0"
	configFile = "config.json"
	cfg        = defaultConfig()
	log        = zaplog.Zap("transfer")
)

func prepare() {
	osutil.Flags(version, BuildTime, cfg)
	log.SetDebug(cfg.Debug)
	log.Info("init", "core", runtime.GOMAXPROCS(0), "version", version)

	// utils.WriteConfigFile(configFile, cfg)
	if err := utils.ReadConfigFile(configFile, &cfg); err != nil {
		log.Fatal("config-error", "error", err)
	}
}

func main() {

	prepare()

	// set center
	configapi.SetAPI(cfg.CenterAddr)
	configapi.SetIntervals([]string{"endpoints", "heartbeat"})
	go configapi.Intervals(time.Second * 20)

	// set token
	go httptoken.SyncVerifier(cfg.TokenFile, time.Second*15)

	// prepare queues
	mQueue := queues.NewQueue(1e6, "_dump/metrics")
	go mQueue.ScanDump(time.Minute, func(res queues.ScanDumpResult) {
		log.Info("read-metrics-dump", "dump", res)
	})
	evtQueue := queues.NewQueue(1e6, "_dump/events")
	go evtQueue.ScanDump(time.Minute, func(res queues.ScanDumpResult) {
		log.Info("read-events-dump", "dump", res)
	})

	// init event-sender
	eventsender.SetURLs(cfg.EventorAddr)
	go eventsender.ProcessQueue(evtQueue, 200)

	// init http server
	transferhandler.SetQueues(mQueue, evtQueue)
	go httputil.Listen(cfg.HTTPAddr, transferhandler.Create(cfg.IsPublic))

	go expvar.PrintAlways("mallard2_eventor_perf", cfg.PerfFile, time.Minute)

	osutil.Wait()

	httputil.Close()
	eventsender.Stop()

	dump(mQueue, evtQueue)

	log.Sync()
}

func dump(mQueue, evtQueue *queues.Queue) {
	file, count, err := mQueue.Dump(1e6 * 2)
	if err != nil {
		log.Warn("metrics-dump-error", "error", err)
	} else {
		log.Info("metrics-dump", "file", file, "count", count)
	}
	file, count, err = evtQueue.Dump(1e6 * 2)
	if err != nil {
		log.Warn("events-dump-error", "error", err)
	} else {
		log.Info("events-dump", "file", file, "count", count)
	}
}
